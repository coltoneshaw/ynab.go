package api_test

import (
	"fmt"

	"github.com/coltoneshaw/ynab.go/api"
)

func ExampleDateFromString() {
	date, _ := api.DateFromString("2020-01-20")
	fmt.Println(date)

	// Output: 2020-01-20 00:00:00 +0000 UTC
}
