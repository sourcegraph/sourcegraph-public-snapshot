// Copyright 2021 CUE Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package adt

type envYield struct {
	comp *Comprehension
	env  *Environment
	id   CloseInfo
	err  *Bottom
}

func (n *nodeContext) insertComprehension(env *Environment, x *Comprehension, ci CloseInfo) {
	n.comprehensions = append(n.comprehensions, envYield{x, env, ci, nil})
}

// injectComprehensions evaluates and inserts comprehensions.
func (n *nodeContext) injectComprehensions(all *[]envYield) (progress bool) {
	ctx := n.ctx

	k := 0
	for i := 0; i < len(*all); i++ {
		d := (*all)[i]

		sa := []*Environment{}
		f := func(env *Environment) {
			sa = append(sa, env)
		}

		if err := ctx.Yield(d.env, d.comp, f); err != nil {
			if err.IsIncomplete() {
				d.err = err
				(*all)[k] = d
				k++
			} else {
				// continue to collect other errors.
				n.addBottom(err)
			}
			continue
		}

		if len(sa) == 0 {
			continue
		}
		id := d.id.SpawnSpan(d.comp.Clauses, ComprehensionSpan)

		n.ctx.nonMonotonicInsertNest++
		for _, env := range sa {
			n.addExprConjunct(Conjunct{env, d.comp.Value, id})
		}
		n.ctx.nonMonotonicInsertNest--
	}

	progress = k < len(*all)

	*all = (*all)[:k]

	return progress
}
