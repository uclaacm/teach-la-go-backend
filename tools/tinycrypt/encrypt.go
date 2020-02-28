package tinycrypt

import (
	//"fmt"
)

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
func F36(b, k uint64) (uint64){
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
			plain = Swap36((l ^ ((F36(r, k) << 18) & 0xFFFFFFFFF)) | r)
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

// MakeHash is a helper funciton that takes a 120 bit id (20 char) and makes a 36 bit id
// Make sure the char is UTF-8
func MakeHash(uid string) (uint64) {
	
	index := uint64(0)
	m_index := uint64(0)
	
	aid := uint64(0)

	//convert their char into 6 bit ints
	for i := 0; i < 6; i++{

		index = CharToInt(rune(uid[i*3]))
		// at this point index can only contain values [0, 61]. 
		// to encode this into full 64 bits, we need to add a value between [0, 2] 

		// get 2 chars 
		mixer := uid[i*3 + 1: i*3 + 3] 

		// intialize index
		m_index = 0
		// for each char in mixer
		for _, m := range mixer{
			//fmt.Printf("%#U", m);
			m_index = m_index + CharToInt(m)
		}
		// now m_index has a value [0, 122]
		// divide by 41 to obtain a value [0, 2]
		// This method ensures that a value in [0, 2] is chosen with roughly equal probability
		index += m_index / 41

		aid = (aid | index) << 6
	}

	aid = aid >> 6

	return aid

}

// CharToInt converts a characetr [a~zA~Z0~9] 
// into a integer [0, 61]

func CharToInt(c rune) (uint64){
	//if this is a number (0 ~ 9)
		if c >= 0x0030 && c <= 0x0039 {
			return uint64(c - 0x0030) 
		} else if c >= 0x0041 && c <= 0x005A{ //else if this is uppercase char
			return uint64(c - 0x0037) 
		} else if c >= 0x0061 && c <= 0x007A{ //else if this is lowercase char
			return uint64(c - 0x003D)
		} else {
			return 0
		}
}

// GenerateWord36 takes a 36 bit unsigned integer and creates 
// human friendly hashes
func GenerateWord36(plain uint64) ([]string){

	id := uint16(0)
	mask := uint64(0xFFF) //12 bits

	words := []string{}
	
	for i := 0; i < 3; i++{

		id = uint16(plain & mask) 
		words = append(words, Words[id])
		plain = plain >> 12
	}

	return words
}

