package tinycrypt

// https://en.wikipedia.org/wiki/Feistel_cipher

type Encrypter struct{
	Key int
}

func (e *Encrypter) InitializeEncrypter(key int) {
	e.Key = key
}

func F(b, k int) (int){
	return (b ^ k) 
}

func (e *Encrypter) Encrypt36(plain int) (int){

		// Get the rightmost 18 bits 
		r := 0x3FFFF & plain
		// get the leftmost 18 bits 
		l := (0xFFFFC0000 & plain) 
		
		return (l ^ (F(r, e.Key) << 18)) | r

}

func F8(b, k uint8) (uint8){
	return (b ^ k) 
}

func (e *Encrypter) Encrypt8(plain uint8) (uint8){

	// Get the rightmost 18 bits 
	r := 0xF & plain
	// get the leftmost 18 bits 
	l := (0xF0 & plain) 
	
	return (l ^ (F8(r, uint8(e.Key)) << 4)) | r

}