package user_test

import (
	"fmt"
	"reflect"

	"github.com/coltoneshaw/ynab.go"
)

func ExampleService_GetUser() {
	c := ynab.NewClient("<valid_ynab_access_token>")
	user, _ := c.User().GetUser()
	fmt.Println(reflect.TypeOf(user))

	// Output: *user.User
}
