//go:build ruleguard

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func forbidIfaceEq(m dsl.Matcher) {
	m.Match(`$x == $y`, `$x != $y`).
		Where(
			// Поддерживает как interface{}, так и именованные интерфейсы, через Underlying().
			m["x"].Type.Underlying().Is(`interface{}`) &&
				m["y"].Type.Underlying().Is(`interface{}`) &&
				!m["x"].Text.Matches(`^nil$`) &&
				!m["y"].Text.Matches(`^nil$`),
		).
		Report(`comparison of interfaces via ==/!= is forbidden: may panic on incomparable dynamic type; use type switch + slices.Equal/maps.Equal/reflect.DeepEqual`).
		At(m["x"])
}
