package perms

import "testing"

func equalButUnsorted(p1, p2 Array) bool {
	for _, v1 := range p1 {
		contains := false
		for _, v2 := range p2 {
			if v1 == v2 {
				contains = true
			}
		}
		if !contains {
			return false
		}
	}

	for _, v2 := range p2 {
		contains := false
		for _, v1 := range p1 {
			if v1 == v2 {
				contains = true
			}
		}
		if !contains {
			return false
		}
	}

	return true
}

func TestUpdate(t *testing.T) {

	p1 := Array{
		"+foo.bar",
		"-foo.baz",
	}

	p2, updated := p1.Update("+foo.foobar", false)
	if !updated {
		t.Error("did not changed")
	}
	if !equalButUnsorted(
		p2,
		Array{
			"+foo.bar",
			"-foo.baz",
			"+foo.foobar",
		},
	) {
		t.Error("unexpected update result")
	}

	p2, updated = p1.Update("-foo.foobar", false)
	if !updated {
		t.Error("did not changed")
	}
	if !equalButUnsorted(
		p2,
		Array{
			"+foo.bar",
			"-foo.baz",
			"-foo.foobar",
		},
	) {
		t.Error("unexpected update result")
	}

	p2, updated = p1.Update("+foo.bar", false)
	if updated {
		t.Error("did changed")
	}
	if !equalButUnsorted(
		p2,
		p1,
	) {
		t.Error("unexpected update result")
	}

	p2, updated = p1.Update("+foo.baz", false)
	if !updated {
		t.Error("did not changed")
	}
	if !equalButUnsorted(
		p2,
		Array{
			"+foo.bar",
		},
	) {
		t.Error("unexpected update result")
	}

	p2, updated = p1.Update("-foo.bar", false)
	if !updated {
		t.Error("did not changed")
	}
	if !equalButUnsorted(
		p2,
		Array{
			"-foo.baz",
		},
	) {
		t.Error("unexpected update result")
	}

	p2, updated = p1.Update("-foo.bar", false)
	if !updated {
		t.Error("did not changed")
	}
	if !equalButUnsorted(
		p2,
		Array{
			"-foo.baz",
		},
	) {
		t.Error("unexpected update result")
	}

	p2, updated = p1.Update("-foo.bar", false)
	if !updated {
		t.Error("did not changed")
	}
	if !equalButUnsorted(
		p2,
		Array{
			"-foo.baz",
		},
	) {
		t.Error("unexpected update result")
	}

	p2, updated = p1.Update("-foo.bar", true)
	if !updated {
		t.Error("did not changed")
	}
	if !equalButUnsorted(
		p2,
		Array{
			"-foo.bar",
			"-foo.baz",
		},
	) {
		t.Error("unexpected update result")
	}

	p2, updated = p1.Update("+foo.baz", true)
	if !updated {
		t.Error("did not changed")
	}
	if !equalButUnsorted(
		p2,
		Array{
			"+foo.bar",
			"+foo.baz",
		},
	) {
		t.Error("unexpected update result")
	}

	p2, updated = p1.Update("-foo.foobar", true)
	if !updated {
		t.Error("did not changed")
	}
	if !equalButUnsorted(
		p2,
		Array{
			"+foo.bar",
			"-foo.baz",
			"-foo.foobar",
		},
	) {
		t.Error("unexpected update result")
	}

	p2, updated = p1.Update("+foo.bar", true)
	if updated {
		t.Error("did changed")
	}
	if !equalButUnsorted(
		p2,
		Array{
			"+foo.bar",
			"-foo.baz",
		},
	) {
		t.Error("unexpected update result")
	}

}

func TestEquals(t *testing.T) {
	p1 := Array{
		"+foo.bar",
		"-foo.baz",
	}
	if !p1.Equals(p1) {
		t.Error("equal arrays have unequal res")
	}

	p1 = Array{}
	if !p1.Equals(p1) {
		t.Error("equal arrays have unequal res")
	}

	p1 = Array{
		"+foo.bar",
		"-foo.baz",
	}
	p2 := Array{
		"-foo.baz",
		"+foo.bar",
	}
	if p1.Equals(p2) {
		t.Error("unequal arrays have equal res")
	}

	p1 = Array{
		"+foo.bar",
		"-foo.baz",
	}
	p2 = Array{
		"-foo.baz",
	}
	if p1.Equals(p2) {
		t.Error("unequal arrays have equal res")
	}
}

func TestHas(t *testing.T) {
	p := Array{
		"+foo.bar",
		"+foo.baz.*",
		"-foo.foobar",
		"-foo.foobar.*",
	}
	if !p.Has("foo.bar") {
		t.Error("check failed")
	}

	if !p.Has("foo.baz.c") {
		t.Error("check failed")
	}

	if p.Has("foo.foobar") {
		t.Error("check failed")
	}

	if p.Has("foo.foobar.c") {
		t.Error("check failed")
	}

	if p.Has("foo.foobar") {
		t.Error("check failed")
	}

	if p.Has("x.y.z") {
		t.Error("check failed")
	}

	if p.Has("x") {
		t.Error("check failed")
	}

	if p.Has("") {
		t.Error("check failed")
	}
}
