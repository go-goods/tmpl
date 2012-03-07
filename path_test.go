package tmpl

import "testing"

func BenchmarkPathStringWith(b *testing.B) {
	pth := pathRootedAt(nil)
	sel := []string{"foo", "bar", "baz"}

	for i := 0; i < b.N; i++ {
		pth.StringWith(sel)
	}
}
