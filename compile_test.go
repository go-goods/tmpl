package tmpl

import "testing"

func TestExecuteListString(t *testing.T) {
	l := executeList{
		nil,
		executeList{nil, nil, nil},
		nil,
	}
	l.Push(nil)
	if l.String() != "[\n\tnil\n\t[\n\t\tnil\n\t\tnil\n\t\tnil\n\t]\n\tnil\n\tnil\n]" {
		t.Error("didn't next right")
	}
}
