package main

import "github.com/nsf/termbox-go"

func setSquare(x, y int, ch rune, fg, bg termbox.Attribute) {
	termbox.SetCell(x*2, y, ch, fg, bg)
	termbox.SetCell(x*2+1, y, ch, fg, bg)
}

func writeStr(x, y int, str string, fg, bg termbox.Attribute) {
	for i, c := range str {
		termbox.SetCell(x+i, y, c, fg, bg)
	}
}

// Combinations returns all size-k subsets of arr, order inside each subset
// matching arr’s order. No two subsets differ only by permutation.
func combinations[T any](arr []T, k int) [][]T {
    n := len(arr)
    if k < 0 || k > n {
        return  make([][]T, 0)
    }
    // compute binomial(n,k) to preallocate
    total := binom(n, k)
    out := make([][]T, 0, total)

    // initial indices 0,1,…,k-1
    idx := make([]int, k)
    for i := 0; i < k; i++ {
        idx[i] = i
    }

    for {
        // emit current combination
        combo := make([]T, k)
        for i, v := range idx {
            combo[i] = arr[v]
        }
        out = append(out, combo)

        // find rightmost index to increment
        i := k - 1
        for i >= 0 && idx[i] == i + n - k {
            i--
        }
        if i < 0 {
            break
        }
        // increment and reset following
        idx[i]++
        for j := i + 1; j < k; j++ {
            idx[j] = idx[j-1] + 1
        }
    }

    return out
}

// binom computes C(n,k) in O(k) time without overflow for moderate n,k.
func binom(n, k int) int {
    if k > n-k {
        k = n - k
    }
    res := 1
    for i := 1; i <= k; i++ {
        res = res * (n - k + i) / i
    }
    return res
}


// Intersection returns the unique elements present in every slice within lists.
// Preserves the order of the first slice.
// T must be a comparable type. T can be a struct if all fields are comparable.
func intersection[T comparable](slices [][]T) []T {
    if len(slices) == 0 {
        return nil
    }
    // count occurrences across distinct elements per list
    counts := make(map[T]int)
    for _, list := range slices {
        seen := make(map[T]struct{}, len(list))
        for _, v := range list {
            if _, ok := seen[v]; !ok {
                seen[v] = struct{}{}
                counts[v]++
            }
        }
    }
    // collect only those seen in every list, in order of the first list
    var result []T
    seenFinal := make(map[T]struct{})
    for _, v := range slices[0] {
        if counts[v] == len(slices) {
            if _, added := seenFinal[v]; !added {
                result = append(result, v)
                seenFinal[v] = struct{}{}
            }
        }
    }
    return result
}

// Union returns the unique elements present in any slice within lists.
// Preserves the order of first appearance across the slices.
func union[T comparable](slices [][]T) []T {
	if len(slices) == 0 {
		return nil
	}
	seen := make(map[T]struct{})
	var result []T
	for _, list := range slices {
		for _, v := range list {
			if _, ok := seen[v]; !ok {
				seen[v] = struct{}{}
				result = append(result, v)
			}
		}
	}
	return result
}


// Difference returns the unique elements that appear in a but not in b.
// Order is preserved as in a.
func difference[T comparable](a, b []T) []T {
	// Build a set of elements present in b.
	inB := make(map[T]struct{}, len(b))
	for _, v := range b {
		inB[v] = struct{}{}
	}

	// Collect elements from a that are not in b.
	seen := make(map[T]struct{})
	var result []T
	for _, v := range a {
		if _, found := inB[v]; !found {
			if _, added := seen[v]; !added {
				seen[v] = struct{}{}
				result = append(result, v)
			}
		}
	}
	return result
}
