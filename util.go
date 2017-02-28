package zen

// assert c is true, else panic with msg
func assert(c bool, msg string) {
	if !c {
		panic(msg)
	}
}

func maxUint8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func longestCommonPrefixIndex(a, b string) int {
	i := 0
	max := minInt(len(a), len(b))
	for i < max && a[i] == b[i] {
		i++
	}
	return i
}
