package category_test

import (
	"fmt"

	"github.com/coltoneshaw/ynab.go/api/category"

	"reflect"

	"github.com/coltoneshaw/ynab.go"
	"github.com/coltoneshaw/ynab.go/api"
)

func ExampleService_GetCategory() {
	client := ynab.NewClient("<valid_ynab_access_token>")
	c, _ := client.Category().GetCategory("<valid_budget_id>", "<valid_category_id>")
	fmt.Println(reflect.TypeOf(c))

	// Output: *category.Category
}

func ExampleService_GetCategories() {
	client := ynab.NewClient("<valid_ynab_access_token>")
	f := &api.Filter{LastKnowledgeOfServer: 10}
	categories, _ := client.Category().GetCategories("<valid_budget_id>", f)
	fmt.Println(reflect.TypeOf(categories))

	// Output: *category.SearchResultSnapshot
}

func ExampleService_GetCategoryForMonth() {
	client := ynab.NewClient("<valid_ynab_access_token>")
	c, _ := client.Category().GetCategoryForMonth("<valid_budget_id>",
		"<valid_category_id>", api.Date{})
	fmt.Println(reflect.TypeOf(c))

	// Output: *category.Category
}

func ExampleService_GetCategoryForCurrentMonth() {
	client := ynab.NewClient("<valid_ynab_access_token>")
	c, _ := client.Category().GetCategoryForCurrentMonth("<valid_budget_id>",
		"<valid_category_id>")
	fmt.Println(reflect.TypeOf(c))

	// Output: *category.Category
}

func ExampleService_UpdateCategoryForMonth() {
	validMonth, _ := api.DateFromString("2018-01-01")
	validPayload := category.PayloadMonthCategory{Budgeted: 1000}

	client := ynab.NewClient("<valid_ynab_access_token>")
	c, _ := client.Category().UpdateCategoryForMonth("<valid_budget_id>",
		"<valid_category_id>", validMonth, validPayload)
	fmt.Println(reflect.TypeOf(c))

	// Output: *category.Category
}

func ExampleService_UpdateCategoryForCurrentMonth() {
	validPayload := category.PayloadMonthCategory{Budgeted: 1000}

	client := ynab.NewClient("<valid_ynab_access_token>")
	c, _ := client.Category().UpdateCategoryForCurrentMonth("<valid_budget_id>",
		"<valid_category_id>", validPayload)
	fmt.Println(reflect.TypeOf(c))

	// Output: *category.Category
}
