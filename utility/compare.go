package utility

func BytesEqual(ba, bb []byte) bool {
	if len(ba) != len(bb) {
		return false
	}

	for i := range ba {
		if ba[i] != bb[i] {
			return false
		}
	}
	return true
}
