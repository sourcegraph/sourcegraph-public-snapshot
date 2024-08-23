// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package query

import (
	"strings"
)

type qnode struct {
	query    string
	children []*qnode
}

// HasOrderedResults checks if a given SQL query returns ordered results.
// This function uses a naive approach of checking the root level query
// ( ignoring subqueries, functions calls, etc ) and checking
// if it contains an ORDER BY clause.
func HasOrderedResults(sql string) bool {
	cleanedQuery := strings.TrimSpace(sql)
	if !strings.HasPrefix(cleanedQuery, "(") {
		cleanedQuery = "(" + cleanedQuery + ")"
	}
	root := &qnode{query: cleanedQuery, children: []*qnode{}}
	curNode := root
	indexStack := []int{}
	nodeStack := []*qnode{}
	for i, c := range cleanedQuery {
		if c == '(' {
			indexStack = append(indexStack, i)
			nextNode := &qnode{children: []*qnode{}}
			curNode.children = append(curNode.children, nextNode)
			nodeStack = append(nodeStack, curNode)
			curNode = nextNode
		}
		if c == ')' {
			if len(indexStack) == 0 {
				// unbalanced
				return false
			}
			start := indexStack[len(indexStack)-1]
			indexStack = indexStack[:len(indexStack)-1]

			curNode.query = cleanedQuery[start+1 : i]

			curNode = nodeStack[len(nodeStack)-1]
			nodeStack = nodeStack[:len(nodeStack)-1]
		}
	}
	curNode = root.children[0]
	q := curNode.query
	for _, c := range curNode.children {
		q = strings.Replace(q, c.query, "", 1)
	}
	return strings.Contains(strings.ToUpper(q), "ORDER BY")
}
