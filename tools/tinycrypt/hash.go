package tinycrypt

// https://en.wikipedia.org/wiki/Feistel_cipher
// Note: Perhaps that simplist way to implement this package will be to use 
// Knuth's Multiplicative Method. However this will require us to have a 128 bit,
// which ends up making implementation complex. 
// This method is more convoluted, but only requires bitwise operations and 64 bit ints

// Declare constants

// structure for Encrypter
type Encrypter struct{
	Key []uint64
}


// InitializeEncrypter is a function to initialize the Encrypter with 
// a list of keys. 
// Input: list of uint64 keys
func (e *Encrypter) InitializeEncrypter(key []uint64) {
	e.Key = key
}

// F is a function that takes a key and data, and generates a hash
// F does not have to be inverseable
func F(b, k uint64) (uint64){
	return b + (k ^ (b << 2)) + (k ^ (b << 4)) + (k ^ (b << 8)) + (k ^ (b << 16)) + (k ^ (b << 32))
}

func Swap36(i uint64) (uint64){
	return ((i & 0x3FFFF) << 18) | (0xFFFFC0000 & i) >> 18 
}

func (e *Encrypter) Encrypt36(plain uint64) (uint64){

		var r, l uint64

		for _, k := range e.Key {
			// Get the rightmost 18 bits 
			r = uint64(0x3FFFF) & plain
			// get the leftmost 18 bits 
			l = (uint64(0xFFFFC0000) & plain) 
			// Use key to perform conversion
			plain = Swap36((l ^ ((F(r, k) << 18) & 0xFFFFFFFFF)) | r)
		}
		// Since we swap N-1 times, we need to "cancel out" the last swap 
		return Swap36(plain)

}

// ======================
// Handlers for encrypting 8 bit data
// ======================

func (e *Encrypter) Encrypt8(plain uint8) (uint8){

	var res uint16

	res = (uint16(plain) * 123) % 256

	return uint8(res)

}

// ======================
// Handlers for creating a plain engligh word hash
// ======================

// GenerateWord36 takes a 36 bit unsigned integer and creates 
// human friendly hashes
func GenerateWord36(plain uint64) ([]string){

	id := uint16(0)
	mask := uint64(0xFFF)

	words := []string{}
	
	for i := 0; i < 3; i++{

		id = uint16(plain & mask) 
		words = append(words, Words[id])
		plain = plain >> 12
	}

	return words
}