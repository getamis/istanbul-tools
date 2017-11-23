package tests

import (
	"reflect"

	"github.com/onsi/ginkgo"
)

/*
TableEntry represents an entry in a table test.  You generally use the `Entry` constructor.
*/
type tableEntry struct {
	Description string
	Parameters  []interface{}
	Pending     bool
	Focused     bool
}

func (t tableEntry) generate(itBody reflect.Value, entries []tableEntry, pending bool, focused bool) {
	if t.Pending {
		ginkgo.PDescribe(t.Description, func() {
			for _, entry := range entries {
				entry.generate(itBody, entries, pending, focused)
			}
		})
		return
	}

	values := []reflect.Value{}
	for i, param := range t.Parameters {
		var value reflect.Value

		if param == nil {
			inType := itBody.Type().In(i)
			value = reflect.Zero(inType)
		} else {
			value = reflect.ValueOf(param)
		}

		values = append(values, value)
	}

	body := func() {
		itBody.Call(values)
	}

	if t.Focused {
		ginkgo.FDescribe(t.Description, body)
	} else {
		ginkgo.Describe(t.Description, body)
	}
}

/*
Entry constructs a tableEntry.

The first argument is a required description (this becomes the content of the generated Ginkgo `It`).
Subsequent parameters are saved off and sent to the callback passed in to `DescribeTable`.

Each Entry ends up generating an individual Ginkgo It.
*/
func Case(description string, parameters ...interface{}) tableEntry {
	return tableEntry{description, parameters, false, false}
}

/*
You can focus a particular entry with FEntry.  This is equivalent to FIt.
*/
func FCase(description string, parameters ...interface{}) tableEntry {
	return tableEntry{description, parameters, false, true}
}

/*
You can mark a particular entry as pending with PEntry.  This is equivalent to PIt.
*/
func PCase(description string, parameters ...interface{}) tableEntry {
	return tableEntry{description, parameters, true, false}
}

/*
You can mark a particular entry as pending with XEntry.  This is equivalent to XIt.
*/
func XCase(description string, parameters ...interface{}) tableEntry {
	return tableEntry{description, parameters, true, false}
}
