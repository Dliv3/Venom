// Modified https://github.com/paddie/gokmp
package utils

import (
	"errors"
	"fmt"
)

type KMP struct {
	Pattern string
	Prefix  []int
	Size    int
}

// For debugging
func (kmp *KMP) String() string {
	return fmt.Sprintf("pattern: %v\nprefix: %v", kmp.Pattern, kmp.Prefix)
}

// compile new prefix-array given argument
func NewKMP(pattern string) (*KMP, error) {
	prefix, err := computePrefix(pattern)
	if err != nil {
		return nil, err
	}
	return &KMP{
			Pattern: pattern,
			Prefix:  prefix,
			Size:    len(pattern)},
		nil
}

// returns an array containing indexes of matches
// - error if pattern argument is less than 1 char
func computePrefix(pattern string) ([]int, error) {
	// sanity check
	len_p := len(pattern)
	if len_p < 2 {
		if len_p == 0 {
			return nil, errors.New("'pattern' must contain at least one character")
		}
		return []int{-1}, nil
	}
	t := make([]int, len_p)
	t[0], t[1] = -1, 0

	pos, count := 2, 0
	for pos < len_p {

		if pattern[pos-1] == pattern[count] {
			count++
			t[pos] = count
			pos++
		} else {
			if count > 0 {
				count = t[count]
			} else {
				t[pos] = 0
				pos++
			}
		}
	}
	return t, nil
}
