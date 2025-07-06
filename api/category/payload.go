// Copyright (c) 2018, Bruno M V Souza <github@b.bmvs.io>. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause license that can be
// found in the LICENSE file.

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
