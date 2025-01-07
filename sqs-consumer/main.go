package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/lib/pq"
)

type SQSMessage struct {
	EmailSource string `json:"email_source"`
	EmailTarget string `json:"email_target"`
	Date        string `json:"date"`
	WantsMatch  bool   `json:"wants_match"`
}

var (
	db     *sql.DB
	dbOnce sync.Once
)

func main() {
	lambda.Start(handleSQSEvent)
}

// remove message on failure? return an error
func handleSQSEvent(ctx context.Context, event events.SQSEvent) error {
	initDB()

	for _, record := range event.Records {
		var payload SQSMessage

		if err := json.Unmarshal([]byte(record.Body), &payload); err != nil {
			log.Printf("Failed to unmarshal SQS message ID=%s: %v", record.MessageId, err)
			continue
		}

		log.Printf("Processing message: Source=%s, Target=%s", payload.EmailSource, payload.EmailTarget)

		if err := insertMatch(ctx, payload); err != nil {
			return fmt.Errorf("insertMatch failed for msg ID=%s: %w", record.MessageId, err)
		}
	}

	return nil
}

func GetEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
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

func insertMatch(ctx context.Context, msg SQSMessage) error {
	parsedDate, err := time.Parse("2006-01-02", msg.Date)
	if err != nil {
		return fmt.Errorf("invalid date '%s': %w", msg.Date, err)
	}

	sundayOfWeek := parsedDate.AddDate(0, 0, -int(parsedDate.Weekday()))

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

	// (source email == user1_email OR user2_email) AND week = sundayOfWeek
	lockQuery := `
      SELECT user1_email, user2_email, server_generated
        FROM matches
       WHERE (user1_email = $1 OR user2_email = $1)
         AND date_trunc('week', week) = date_trunc('week', $2::timestamp)
       FOR UPDATE
    `
	lockRows, lockErr := tx.QueryContext(ctx, lockQuery, msg.EmailSource, sundayOfWeek)
	if lockErr != nil {
		return fmt.Errorf("locking rows failed: %w", lockErr)
	}
	defer lockRows.Close()

	// var serverGeneratedPairs []string
	var notServerGeneratedButNotSource []string
	var countNonServer int
	var targetMatchIsServerGenerated bool
	// var targetEmail string
	var isTarget1 bool

	for lockRows.Next() {
		var u1, u2 string
		var sg bool
		if scanErr := lockRows.Scan(&u1, &u2, &sg); scanErr != nil {
			return fmt.Errorf("scan locked rows failed: %w", scanErr)
		}

		if sg {
			if u1 == msg.EmailTarget {
				targetMatchIsServerGenerated = true
				// targetEmail = u1
				isTarget1 = true
				break
			} else if u2 == msg.EmailTarget {
				targetMatchIsServerGenerated = true
				// targetEmail = u2
				isTarget1 = false
				break
			}
			// serverGeneratedPairs = append(
			// serverGeneratedPairs,
			// fmt.Sprintf("(%s, %s)", u1, u2),
			// )
		} else {
			// record ones which do not include EmailSource
			if u1 == msg.EmailTarget {
				notServerGeneratedButNotSource = append(
					notServerGeneratedButNotSource,
					fmt.Sprintf("(%s, %s)", u1, u2),
				)
				isTarget1 = true
			} else if u2 == msg.EmailTarget {
				notServerGeneratedButNotSource = append(
					notServerGeneratedButNotSource,
					fmt.Sprintf("(%s, %s)", u1, u2),
				)
				isTarget1 = false
			}
			countNonServer++
		}
	}
	if targetMatchIsServerGenerated {
		// update row to set wants_match to val
		if isTarget1 {
			updateSQL := `
                UPDATE matches SET user2_interested = $1 WHERE user1_email = $2 AND user2_email = $3 AND week = $4
                `
			if _, updErr := tx.ExecContext(ctx, updateSQL, msg.WantsMatch, msg.EmailTarget, msg.EmailSource, sundayOfWeek); updErr != nil {
				return fmt.Errorf("update failed: %w", updErr)
			}
		} else {
			updateSQL := `
                UPDATE matches SET user1_interested = $1 WHERE user1_email = $2 AND user2_email = $3 AND week = $4
            `
			if _, updErr := tx.ExecContext(ctx, updateSQL, msg.WantsMatch, msg.EmailSource, msg.EmailTarget, sundayOfWeek); updErr != nil {
				return fmt.Errorf("update failed: %w", updErr)
			}
		}
		return nil
	}

	// target is not server generated
	if countNonServer > 0 {
		// check if email_target is already in a insertMatch(
		if countNonServer > 1 || len(notServerGeneratedButNotSource) == 0 {
			// can't insert match b/c user already picked one match
			return nil
		}
		// update row in table to set wants match to val
		if isTarget1 {
			updateSQL := `
                UPDATE matches SET user2_interested = $1 WHERE user1_email = $2 AND user2_email = $3 AND week = $4
            `
			if _, updErr := tx.ExecContext(ctx, updateSQL, msg.WantsMatch, msg.EmailTarget, msg.EmailSource, sundayOfWeek); updErr != nil {
				return fmt.Errorf("update failed: %w", updErr)
			}
		} else {
			updateSQL := `
                UPDATE matches SET user1_interested = $1 WHERE user1_email = $2 AND user2_email = $3 AND week = $4
            `
			if _, updErr := tx.ExecContext(ctx, updateSQL, msg.WantsMatch, msg.EmailSource, msg.EmailTarget, sundayOfWeek); updErr != nil {
				return fmt.Errorf("update failed: %w", updErr)
			}
		}

		return nil
	}

	// insert row into table
	insertSQL := `
      INSERT INTO matches (user1_email, user2_email, user1_interested, week)
           VALUES ($1,        $2,          $3,              $4)
    `
	if _, insErr := tx.ExecContext(ctx, insertSQL, msg.EmailSource, msg.EmailTarget, true, sundayOfWeek); insErr != nil {
		return fmt.Errorf("insert failed: %w", insErr)
	}

	return nil
}
