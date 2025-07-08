// Copyright (c) 2019, Bruno M V Souza <github@b.bmvs.io>. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause license that can be
// found in the LICENSE file.

package api_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/coltoneshaw/ynab.go/api"
)

func TestFilter_ToQuery(t *testing.T) {
	table := []struct {
		Name   string
		Input  api.Filter
		Output string
	}{
		{
			Name:   "with_server_knowledge",
			Input:  api.Filter{LastKnowledgeOfServer: 2},
			Output: "last_knowledge_of_server=2",
		},
		{
			Name:   "zero_server_knowledge",
			Input:  api.Filter{LastKnowledgeOfServer: 0},
			Output: "last_knowledge_of_server=0",
		},
		{
			Name:   "empty_filter",
			Input:  api.Filter{},
			Output: "last_knowledge_of_server=0",
		},
		{
			Name:   "large_server_knowledge",
			Input:  api.Filter{LastKnowledgeOfServer: 9999999999},
			Output: "last_knowledge_of_server=9999999999",
		},
		{
			Name:   "max_uint64",
			Input:  api.Filter{LastKnowledgeOfServer: ^uint64(0)},
			Output: "last_knowledge_of_server=18446744073709551615",
		},
	}

	for _, test := range table {
		t.Run(test.Name, func(t *testing.T) {
			assert.Equal(t, test.Output, test.Input.ToQuery())
		})
	}
}
