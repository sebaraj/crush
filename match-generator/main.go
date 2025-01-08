package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/lib/pq"
)

var (
	db     *sql.DB
	dbOnce sync.Once
)

func main() {
	lambda.Start(handleMatchGen)
}

const (
	answerCount = 12
	capacity    = 3
)

func handleMatchGen(ctx context.Context, event events.SQSEvent) error {
	initDB()

	// get count of active users
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}

	// handle panic and rollback on complex transactions
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	countQuery := `
        SELECT COUNT(DISTINCT email) FROM users WHERE is_active = true
    `
	var freq int
	row := tx.QueryRowContext(ctx, countQuery)
	countErr := row.Scan(&freq)
	if countErr != nil {
		return fmt.Errorf("locking rows failed: %w", countErr)
	}
	// init lists
	emailList := make([]string, freq)
	myGenderList := make([]int, freq)
	targetGenderList := make([]int, freq)
	answersList := make([][]int, freq)
	for i := range answersList {
		answersList[i] = make([]int, answerCount)
	}
	distanceList := make([][]int, freq)
	for i := range distanceList {
		distanceList[i] = make([]int, freq)
	}

	// SQL query to popular lists
	selectQuery := `
        SELECT 
			u.email, 
			u.gender, 
			u.partner_genders, 
			a.question1, 
			a.question2, 
			a.question3, 
			a.question4, 
			a.question5, 
			a.question6, 
			a.question7, 
			a.question8, 
			a.question9, 
			a.question10, 
			a.question11, 
			a.question12
		FROM users u
		LEFT JOIN answers a ON u.email = a.email
		WHERE u.is_active = true
	`

	rows, err := tx.QueryContext(ctx, selectQuery)
	if err != nil {
		return fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()
	idx := 0
	var email string
	var gender int
	var partnerGenders int
	answers := make([]int, answerCount)

	// weights for manhattan distances
	weightsMul := [answerCount]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	weightsAdd := [answerCount]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	for rows.Next() {
		if scanErr := rows.Scan(&email, &gender, &partnerGenders, &answers[0], &answers[1], &answers[2], &answers[3], &answers[4], &answers[5], &answers[6], &answers[7], &answers[8], &answers[9], &answers[10], &answers[11]); scanErr != nil {
			continue
		}
		emailList[idx] = email
		myGenderList[idx] = gender
		targetGenderList[idx] = partnerGenders
		tempAns := make([]int, answerCount)
		copy(tempAns, answers)
		answersList[idx] = tempAns
		idx++
	}

	// calculate weighted manhattan distances
	for i := 0; i < freq-1; i++ {
		for j := i + 1; j < freq; j++ {
			dist := 0
			// check if genders are compatible
			if i == j || ((myGenderList[i]&targetGenderList[j]) == 0 || (myGenderList[j]&targetGenderList[i]) == 0) {
				// if not compatible, set distance to max
				dist = math.MaxInt
			} else {
				for k := 0; k < answerCount; k++ {
					temp := answersList[i][k] - answersList[j][k]
					if temp < 0 {
						temp = -temp // go please add abs() for ints and not just float64! :(
					}
					dist += weightsMul[k]*temp + weightsAdd[k]
				}
			}
			// store dist in [i,j] and [j,i]
			distanceList[i][j] = dist
			distanceList[j][i] = dist

		}
	}

	preferenceLists := make([][]int, freq)
	for i := 0; i < freq; i++ {
		var candidates []int
		for j := 0; j < freq; j++ {
			if i != j && distanceList[i][j] != math.MaxInt {
				candidates = append(candidates, j)
			}
		}
		sortByDist := func(a, b int) bool {
			return distanceList[i][a] < distanceList[i][b]
		}
		sort.Slice(candidates, sortByDist)

		preferenceLists[i] = candidates
	}

	// Gale-Shapley variant for top 3 matches
	// Data structures for the matching algorithm:
	//  - nextChoice[i] = index into preferenceLists[i], telling whom i will propose to next
	//  - matchedWith[i] = slice of users that have tentatively accepted i (size <= capacity)
	//  - We also need a quick way for the “acceptor” to compare two proposers:
	//    rank[i][j] = how user i ranks user j (lower = more preferred)
	nextChoice := make([]int, freq)
	matchedWith := make([][]int, freq)
	for i := 0; i < freq; i++ {
		matchedWith[i] = make([]int, 0, capacity)
	}

	// rank[i][j]: position of j in i's preference list (lower = more preferred)
	rank := make([][]int, freq)
	for i := 0; i < freq; i++ {
		rank[i] = make([]int, freq)
		// For each j in preferenceLists[i], assign rank
		for pos, userJ := range preferenceLists[i] {
			rank[i][userJ] = pos
		}
	}
	// Function to find the "worst" matched partner for user u (i.e. highest rank)
	// returns that partner's index and their rank
	worstPartner := func(u int) (partner int, partnerRank int) {
		worst := -1
		worstPos := -1
		for _, p := range matchedWith[u] {
			rk := rank[u][p]
			if rk > worstPos {
				worstPos = rk
				worst = p
			}
		}
		return worst, worstPos
	}

	for {
		changed := false

		// Try each user if they still have room for matches < capacity (or have proposals left)
		for i := 0; i < freq; i++ {
			// If i is already at capacity, skip
			if len(matchedWith[i]) >= capacity {
				continue
			}

			// If i has exhausted all possible proposals, skip
			if nextChoice[i] >= len(preferenceLists[i]) {
				continue
			}

			// Propose to the next candidate on i's preference list
			proposeTo := preferenceLists[i][nextChoice[i]]
			nextChoice[i]++

			// If proposeTo is not at capacity, they accept i
			if len(matchedWith[proposeTo]) < capacity {
				matchedWith[proposeTo] = append(matchedWith[proposeTo], i)
				matchedWith[i] = append(matchedWith[i], proposeTo)
				changed = true

			} else {
				// proposeTo is at capacity => check if i is better than their worst
				worstP, worstRank := worstPartner(proposeTo)
				myRank := rank[proposeTo][i]

				if myRank < worstRank {
					// proposeTo drops worstP, accepts i
					matchedWith[proposeTo] = removeOne(matchedWith[proposeTo], worstP)
					matchedWith[worstP] = removeOne(matchedWith[worstP], proposeTo)

					matchedWith[proposeTo] = append(matchedWith[proposeTo], i)
					matchedWith[i] = append(matchedWith[i], proposeTo)
					changed = true

				}
				//else {
				// proposeTo rejects i => do nothing, i remains unmatched
				//}
			}
		}

		if !changed {
			// No change => stable
			break
		}
	}

	// store matches in DB
	sunday := getThisWeeksSunday() // this weeks sunday
	insertStmt := `INSERT INTO matches (email1, email2, server_generated, week) VALUES ($1, $2, $3, $4)`
	for i := 0; i < freq; i++ {
		for _, partner := range matchedWith[i] {
			if i < partner {
				if _, iErr := tx.ExecContext(ctx, insertStmt, emailList[i], emailList[partner], true, sunday); iErr != nil {
					return fmt.Errorf("failed to insert match (%s, %s): %w",
						emailList[i], emailList[partner], iErr)
				}
			}
		}
	}

	return nil
}

func getThisWeeksSunday() time.Time {
	now := time.Now()
	offset := (int(now.Weekday()) + 7 - int(time.Sunday)) % 7
	sunday := time.Date(
		now.Year(),
		now.Month(),
		now.Day()-offset,
		0, 0, 0, 0,
		now.Location(),
	)
	return sunday
}

func GetEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func removeOne(s []int, x int) []int {
	n := len(s)
	for i, val := range s {
		if val == x {
			s[i], s[n-1] = s[n-1], s[i]
			return s[:n-1]
		}
	}
	return s
}

func initDB() {
	dbOnce.Do(func() {
		dbUser := GetEnv("DB_USERNAME", "localtest")
		dbPassword := GetEnv("DB_PASSWORD", "localtest")
		dbEndpoint := GetEnv("DB_ENDPOINT", "localhost")
		dbPort := GetEnv("DB_PORT", "5432")
		dbName := GetEnv("DB_NAME", "my_database")
		if dbUser == "" || dbPassword == "" || dbEndpoint == "" || dbPort == "" || dbName == "" {
			log.Fatal("One or more required environment variables are missing")
		}
		dbEndpoint = strings.Split(dbEndpoint, ":")[0]

		ips, err := net.LookupIP(dbEndpoint)
		if err != nil {
			log.Fatalf("Failed to resolve hostname: %v", err)
		}

		if len(ips) == 0 {
			log.Fatalf("No IP addresses found for hostname: %s", dbEndpoint)
		}

		dbIP := ips[0].String()
		psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
			dbIP, dbPort, dbUser, dbPassword, dbName)

		db, err := sql.Open("postgres", psqlInfo)
		if err != nil {
			log.Fatalf("Unable to connect to DB: %v", err)
		}

		err = db.Ping()
		if err != nil {
			log.Fatalf("Failed to ping DB: %v", err)
		}
	})
}
