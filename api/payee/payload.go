// Copyright (c) 2018, Bruno M V Souza <github@b.bmvs.io>. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause license that can be
// found in the LICENSE file.

package payee

// PayloadPayee represents the payload for updating a payee
type PayloadPayee struct {
	Name string `json:"name"`
}
