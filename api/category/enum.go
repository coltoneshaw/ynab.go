package category

// Goal represents the goal of a category
type Goal string

// Pointer returns the pointer of a Goal
func (g Goal) Pointer() *Goal {
	return &g
}

const (
	// GoalTargetCategoryBalance Goal targets category balance
	GoalTargetCategoryBalance Goal = "TB"
	// GoalTargetCategoryBalanceByDate Goal targets category balance by date
	GoalTargetCategoryBalanceByDate Goal = "TBD"
	// GoalMonthlyFunding Goal by monthly funding
	GoalMonthlyFunding Goal = "MF"
	// GoalPlanYourSpending Plan Your Spending goal
	GoalPlanYourSpending Goal = "NEED"
	// GoalDebt Debt goal
	GoalDebt Goal = "DEBT"
)
