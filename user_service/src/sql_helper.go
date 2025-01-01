package main

import (
	"database/sql"
)

func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func getIntValue(ni sql.NullInt64) int {
	if ni.Valid {
		return int(ni.Int64)
	}
	return 0
}

func filterNullStrings(nullStrings []sql.NullString) []string {
	var result []string
	for _, ns := range nullStrings {
		if ns.Valid {
			result = append(result, ns.String)
		} else {
			result = append(result, "")
		}
	}
	return result
}

func filterNullInts(nullInts []sql.NullInt64) []int {
	var result []int
	for _, ni := range nullInts {
		if ni.Valid {
			result = append(result, int(ni.Int64))
		} else {
			result = append(result, -1)
		}
	}
	return result
}

func joinFields(fields []string, delimiter string) string {
	result := ""
	for i, field := range fields {
		if i > 0 {
			result += delimiter + " "
		}
		result += field
	}
	return result
}
