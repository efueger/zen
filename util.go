package zen

func assert(c bool, msg string) {
	if !c {
		panic(msg)
	}
}
