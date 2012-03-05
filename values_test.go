package tmpl

import (
	"reflect"
	"testing"
)

func TestValueParseSelector(t *testing.T) {
	cases := []struct {
		name string
		tmpl string
		sel  *selectorValue
	}{
		{`basic`, `{% .foo.bar %}`, &selectorValue{0, false, []string{"foo", "bar"}}},
		{`rooted`, `{% /.foo.bar %}`, &selectorValue{0, true, []string{"foo", "bar"}}},
		{`relative`, `{% $$.foo.bar %}`, &selectorValue{2, false, []string{"foo", "bar"}}},
		{`empty`, `{% . %}`, &selectorValue{0, false, nil}},
	}

	for _, c := range cases {
		tree, err := parse(lex([]byte(c.tmpl)))
		if err != nil {
			t.Errorf("%s: failed to parse: %s", c.name, err)
			continue
		}

		if !reflect.DeepEqual(tree.base, c.sel) {
			t.Errorf("%s: not equal:\n%v\n%v", c.name, tree, c.sel)
		}
	}
}
