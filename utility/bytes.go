package utility

func WipeBytes(bs []byte) {
	for i := 0; i < len(bs); i++ {
		bs[i] = 0
	}
}

func WipeBytesOnErr(bs []byte, err *error) {
	if *err != nil {
		for i := 0; i < len(bs); i++ {
			bs[i] = 0
		}
	}
}
