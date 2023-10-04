package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

func getMonthList() []string {
	// Defina os nomes dos meses em português em letras minúsculas
	months := []string{
		"janeiro",
		"fevereiro",
		"março",
		"abril",
		"maio",
		"junho",
		"julho",
		"agosto",
		"setembro",
		"outubro",
		"novembro",
		"dezembro",
	}
	return months
}

func containsMonth(date string) (string, int) {
	months := getMonthList()
	for index, month := range months {
		if strings.Contains(date, month) {
			return month, index
		}
	}
	return "", 0
}

func containsDayBeforeMonth(input string, month string) string {
	index := strings.Index(input, month)
	regex := regexp.MustCompile(`\d{2}`)
	matches := regex.FindAllString(input[:index], -1)
	if matches != nil && len(matches) > 0 {
		lastMatch := matches[len(matches)-1]
		return lastMatch
	}

	return ""
}

func containsYearAfterMonth(input string, month string) string {
	regex := regexp.MustCompile(`\d{4}`)
	index := strings.Index(input, month)
	match := regex.FindString(input[index+len(month):])
	return match
}

// This func must be Exported, Capitalized, and comment added.
// date = Sábado, 25 de Novembro de 2023 - Abertura: 20:00
func ParseDate(date string) (*time.Time, error) {
	regex := regexp.MustCompile(`\d{2}/\d{2}/\d{4}`)
	match := regex.FindString(date)
	layout := "02/01/2006"
	if match != "" {
		parsedTime, err := time.Parse(layout, date)
		if err != nil {
			return nil, err
		}
		return &parsedTime, nil
	}

	month, monthIndex := containsMonth(strings.ToLower(date))

	if month != "" {
		year := containsYearAfterMonth(strings.ToLower(date), month)
		day := containsDayBeforeMonth(strings.ToLower(date), month)
		stringDate := day + "/" + fmt.Sprintf("%02d", (monthIndex+1)) + "/" + year
		parsedTime, err := time.Parse(layout, stringDate)
		if err != nil {
			return nil, err
		}
		return &parsedTime, nil
	}

	return nil, nil
}
