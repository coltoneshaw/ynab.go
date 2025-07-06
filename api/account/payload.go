// Copyright (c) 2018, Bruno M V Souza <github@b.bmvs.io>. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause license that can be
// found in the LICENSE file.

package account

// PayloadAccount represents the payload for creating an account
type PayloadAccount struct {
	Name    string `json:"name"`
	Type    Type   `json:"type"`
	Balance int64  `json:"balance"`
}
