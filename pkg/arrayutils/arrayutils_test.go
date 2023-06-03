package arrayutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexOf(t *testing.T) {

	intArr := []int{1, 2, 3, 4, 5}
	strArr := []string{"a", "b", "c", "d", "e"}

	assert.Equal(t, 0, IndexOf(intArr, 1))
	assert.Equal(t, 1, IndexOf(intArr, 2))

	assert.Equal(t, 0, IndexOf(strArr, "a"))
	assert.Equal(t, 1, IndexOf(strArr, "b"))

}

func TestRemove(t *testing.T) {

	intArr := []int{1, 2, 3, 4, 5}
	strArr := []string{"a", "b", "c", "d", "e"}

	assert.Equal(t, []int{2, 3, 4, 5}, Remove(intArr, 0))
	assert.Equal(t, []int{1, 3, 4, 5}, Remove(intArr, 1))
	assert.Equal(t, []int{1, 2, 4, 5}, Remove(intArr, 2))
	assert.Equal(t, []int{1, 2, 3, 5}, Remove(intArr, 3))
	assert.Equal(t, []int{1, 2, 3, 4}, Remove(intArr, 4))

	assert.Equal(t, []string{"b", "c", "d", "e"}, Remove(strArr, 0))
	assert.Equal(t, []string{"a", "c", "d", "e"}, Remove(strArr, 1))
	assert.Equal(t, []string{"a", "b", "d", "e"}, Remove(strArr, 2))
	assert.Equal(t, []string{"a", "b", "c", "e"}, Remove(strArr, 3))
	assert.Equal(t, []string{"a", "b", "c", "d"}, Remove(strArr, 4))

}

func TestRemoveLazy(t *testing.T) {

	intArr := []int{1, 2, 3, 4, 5}
	strArr := []string{"a", "b", "c", "d", "e"}

	assert.Equal(t, []int{2, 3, 4, 5}, RemoveLazy(intArr, 1))
	assert.Equal(t, []int{1, 3, 4, 5}, RemoveLazy(intArr, 2))
	assert.Equal(t, []int{1, 2, 4, 5}, RemoveLazy(intArr, 3))
	assert.Equal(t, []int{1, 2, 3, 5}, RemoveLazy(intArr, 4))
	assert.Equal(t, []int{1, 2, 3, 4}, RemoveLazy(intArr, 5))

	assert.Equal(t, []string{"b", "c", "d", "e"}, RemoveLazy(strArr, "a"))
	assert.Equal(t, []string{"a", "c", "d", "e"}, RemoveLazy(strArr, "b"))
	assert.Equal(t, []string{"a", "b", "d", "e"}, RemoveLazy(strArr, "c"))
	assert.Equal(t, []string{"a", "b", "c", "e"}, RemoveLazy(strArr, "d"))
	assert.Equal(t, []string{"a", "b", "c", "d"}, RemoveLazy(strArr, "e"))

}

func TestAdd(t *testing.T) {

	intArr := []int{1, 2, 3, 4, 5}
	strArr := []string{"a", "b", "c", "d", "e"}

	assert.Equal(t, []int{0, 1, 2, 3, 4, 5}, Add(intArr, 0, 0))
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, Add(intArr, 6, -1))
	assert.Equal(t, []int{1, 2, 3, 4, 5, 4}, Add(intArr, 4, 5))
	assert.Equal(t, []int{1, 7, 2, 3, 4, 5}, Add(intArr, 7, 1))
	assert.Equal(t, []int{1, 2, 3, 4, 8, 5}, Add(intArr, 8, 4))

	assert.Equal(t, []string{"x", "a", "b", "c", "d", "e"}, Add(strArr, "x", 0))
	assert.Equal(t, []string{"a", "b", "c", "d", "e", "f"}, Add(strArr, "f", -1))
	assert.Equal(t, []string{"a", "b", "c", "d", "e", "d"}, Add(strArr, "d", 5))
	assert.Equal(t, []string{"a", "z", "b", "c", "d", "e"}, Add(strArr, "z", 1))
	assert.Equal(t, []string{"a", "b", "c", "d", "y", "e"}, Add(strArr, "y", 4))

}

func TestContains(t *testing.T) {

	intArr := []int{1, 2, 3, 4, 5}
	strArr := []string{"a", "b", "c", "d", "e"}

	assert.True(t, Contains(intArr, 1))
	assert.True(t, Contains(intArr, 2))
	assert.True(t, Contains(intArr, 3))
	assert.True(t, Contains(intArr, 4))
	assert.True(t, Contains(intArr, 5))

	assert.True(t, Contains(strArr, "a"))
	assert.True(t, Contains(strArr, "b"))
	assert.True(t, Contains(strArr, "c"))
	assert.True(t, Contains(strArr, "d"))
	assert.True(t, Contains(strArr, "e"))

}

func TestContained(t *testing.T) {

	intArr := []int{1, 2, 3, 4, 5}
	strArr := []string{"a", "b", "c", "d", "e"}

	subsetInt1 := []int{2, 4}
	subsetInt2 := []int{6, 7}
	subsetInt3 := []int{1, 6}
	subsetInt4 := []int{3, 5}
	subsetInt5 := []int{0, 3, 5, 9}

	subsetStr1 := []string{"a", "c"}
	subsetStr2 := []string{"f", "g"}
	subsetStr3 := []string{"b", "h"}
	subsetStr4 := []string{"d", "e"}
	subsetStr5 := []string{"x", "c", "e", "z"}

	assert.Equal(t, []int{2, 4}, Contained(subsetInt1, intArr))
	assert.Empty(t, Contained(subsetInt2, intArr))
	assert.Equal(t, []int{1}, Contained(subsetInt3, intArr))
	assert.Equal(t, []int{3, 5}, Contained(subsetInt4, intArr))
	assert.Equal(t, []int{3, 5}, Contained(subsetInt5, intArr))

	assert.Equal(t, []string{"a", "c"}, Contained(subsetStr1, strArr))
	assert.Empty(t, Contained(subsetStr2, strArr))
	assert.Equal(t, []string{"b"}, Contained(subsetStr3, strArr))
	assert.Equal(t, []string{"d", "e"}, Contained(subsetStr4, strArr))
	assert.Equal(t, []string{"c", "e"}, Contained(subsetStr5, strArr))

}

func TestContainsAny(t *testing.T) {

	intArr := []int{1, 2, 3, 4, 5}
	strArr := []string{"a", "b", "c", "d", "e"}

	assert.Equal(t, true, ContainsAny(intArr, 2, 4))
	assert.Equal(t, false, ContainsAny(intArr, 6, 7))
	assert.Equal(t, true, ContainsAny(intArr, 1, 6))
	assert.Equal(t, true, ContainsAny(intArr, 3, 5))
	assert.Equal(t, true, ContainsAny(intArr, 0, 3, 5, 9))

	assert.Equal(t, true, ContainsAny(strArr, "a", "c"))
	assert.Equal(t, false, ContainsAny(strArr, "f", "g"))
	assert.Equal(t, true, ContainsAny(strArr, "b", "h"))
	assert.Equal(t, true, ContainsAny(strArr, "d", "e"))
	assert.Equal(t, true, ContainsAny(strArr, "x", "c", "e", "z"))

}
