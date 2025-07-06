package category_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/jarcoal/httpmock.v1"

	"github.com/brunomvsouza/ynab.go"
	"github.com/brunomvsouza/ynab.go/api"
	"github.com/brunomvsouza/ynab.go/api/category"
)

func TestService_UpdateCategory(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	newGoalTarget := int64(25000)
	newName := "Updated MasterCard"
	payload := category.PayloadCategory{
		Name:       &newName,
		GoalTarget: &newGoalTarget,
	}

	url := "https://api.youneedabudget.com/v1/budgets/aa248caa-eed7-4575-a990-717386438d2c/categories/13419c12-78d3-4a26-82ca-1cde7aa1d6f8"
	httpmock.RegisterResponder(http.MethodPatch, url,
		func(req *http.Request) (*http.Response, error) {
			resModel := struct {
				Category *category.PayloadCategory `json:"category"`
			}{}
			err := json.NewDecoder(req.Body).Decode(&resModel)
			assert.NoError(t, err)
			assert.Equal(t, &payload, resModel.Category)

			res := httpmock.NewStringResponse(200, `{
  "data": {
    "category": {
		"id": "13419c12-78d3-4a26-82ca-1cde7aa1d6f8",
		"category_group_id": "13419c12-78d3-4818-a5dc-601b2b8a6064",
		"name": "Updated MasterCard",
		"hidden": false,
		"original_category_group_id": null,
		"note": null,
		"budgeted": 0,
		"activity": 12190,
		"balance": 18740,
		"deleted": false,
		"goal_type": "TB",
		"goal_creation_month": "2018-04-01",
		"goal_target": 25000,
		"goal_target_month": "2018-05-01",
		"goal_percentage_complete": 20
    }
	}
}
		`)
			return res, nil
		},
	)

	client := ynab.NewClient("")
	c, err := client.Category().UpdateCategory(
		"aa248caa-eed7-4575-a990-717386438d2c",
		"13419c12-78d3-4a26-82ca-1cde7aa1d6f8",
		payload,
	)
	assert.NoError(t, err)

	var (
		expectedGoalTarget             int64 = 25000
		expectedGoalPercentageComplete int32 = 20
	)
	expectedGoalCreationMonth, err := api.DateFromString("2018-04-01")
	assert.NoError(t, err)
	expectedGoalTargetMonth, err := api.DateFromString("2018-05-01")
	assert.NoError(t, err)

	expected := &category.Category{
		ID:                     "13419c12-78d3-4a26-82ca-1cde7aa1d6f8",
		CategoryGroupID:        "13419c12-78d3-4818-a5dc-601b2b8a6064",
		Name:                   "Updated MasterCard",
		Hidden:                 false,
		Budgeted:               int64(0),
		Activity:               int64(12190),
		Balance:                int64(18740),
		Deleted:                false,
		GoalType:               category.GoalTargetCategoryBalance.Pointer(),
		GoalCreationMonth:      &expectedGoalCreationMonth,
		GoalTargetMonth:        &expectedGoalTargetMonth,
		GoalTarget:             &expectedGoalTarget,
		GoalPercentageComplete: &expectedGoalPercentageComplete,
	}
	assert.Equal(t, expected, c)
}
