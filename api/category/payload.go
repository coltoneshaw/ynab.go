package category

// PayloadMonthCategory is the payload contract for updating a category for a month
type PayloadMonthCategory struct {
	Budgeted int64 `json:"budgeted"`
}

// PayloadCategory is the payload contract for updating a category
type PayloadCategory struct {
	Name            *string `json:"name,omitempty"`
	Note            *string `json:"note,omitempty"`
	CategoryGroupID *string `json:"category_group_id,omitempty"`
	GoalTarget      *int64  `json:"goal_target,omitempty"`
}
