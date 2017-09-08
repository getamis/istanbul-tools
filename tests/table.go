// Copyright 2017 AMIS Technologies
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package tests

import (
	"fmt"
	"reflect"

	"github.com/onsi/ginkgo"
)

func CaseTable(description string, itBody interface{}, entries ...tableEntry) bool {
	caseTable(description, itBody, entries, false, false)
	return true
}

/*
You can focus a table with `FCaseTable`.
*/
func FCaseTable(description string, itBody interface{}, entries ...tableEntry) bool {
	caseTable(description, itBody, entries, false, true)
	return true
}

/*
You can mark a table as pending with `PCaseTable`.
*/
func PCaseTable(description string, itBody interface{}, entries ...tableEntry) bool {
	caseTable(description, itBody, entries, true, false)
	return true
}

/*
You can mark a table as pending with `XCaseTable`.
*/
func XCaseTable(description string, itBody interface{}, entries ...tableEntry) bool {
	caseTable(description, itBody, entries, true, false)
	return true
}

func caseTable(description string, itBody interface{}, entries []tableEntry, pending bool, focused bool) {
	itBodyValue := reflect.ValueOf(itBody)
	if itBodyValue.Kind() != reflect.Func {
		panic(fmt.Sprintf("DescribeTable expects a function, got %#v", itBody))
	}

	if pending {
		ginkgo.PDescribe(description, func() {
			for _, entry := range entries {
				entry.generate(itBodyValue, entries, pending, focused)
			}
		})
	} else if focused {
		ginkgo.FDescribe(description, func() {
			for _, entry := range entries {
				entry.generate(itBodyValue, entries, pending, focused)
			}
		})
	} else {
		ginkgo.Describe(description, func() {
			for _, entry := range entries {
				entry.generate(itBodyValue, entries, pending, focused)
			}
		})
	}
}
