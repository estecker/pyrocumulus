package common

import "time"

// https://brandur.org/fragments/go-days-in-month
// daysIn returns the number of days in a given month.
func daysIn(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func PercentageOfMonth() int {
	// Get the current time
	now := time.Now()

	// Get the current year
	year := now.Year()

	// Get the current month
	month := now.Month()

	// Get the number of days in the current month
	days := daysIn(month, year)

	// Get the current day
	day := now.Day()

	// Calculate the percentage of the month that has passed
	percentage := (day * 100) / days

	return percentage
}
