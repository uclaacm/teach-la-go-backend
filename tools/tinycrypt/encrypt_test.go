package tinycrypt_test

import (
	"testing"
	"../tinycrypt"
)

func TestEncrypt(t *testing.T) {

	// var en tinycrypt.Encrypter
	// var de tinycrypt.Encrypter

	// keys := []uint64 {
	// 	0x33684192D,
	// 	0x28DAB6A5A,
	// }

	// keys_rev := []uint64 {
	// 	0x28DAB6A5A,
	// 	0x33684192D,
	// }


	// en.InitializeEncrypter(keys)
	// de.InitializeEncrypter(keys_rev)

	// var res uint64
	// res = 0	

	// t.Logf("Encrypting...")
	
	// for i := 1; i < 256; i++{
	// 	res = uint64(i)
	// 	t.Run("Encrypt", func(t *testing.T){
		
	// 		res = en.Encrypt36(res)
	// 		t.Log("Encrypted: ", res)

	// 		words := tinycrypt.GenerateWord36(res)
	// 		t.Log("Generated: ", words)

	// 		res = de.Encrypt36(res)
	// 		t.Log("Decrypted: ", res)

			
	// 	})
		
	// }

	// var r2 uint8 
	// r2 = 0
	// t.Run("Encrypt loop", func(t *testing.T){
	// 	for i := 0; i < 256; i++{
	// 		r2 = uint8(i)
	// 		//t.Log("Input: ", r2)
	// 		r2 = en.Encrypt8(r2)
	// 		t.Log("", r2)
	// 		r2 = de.Encrypt8(r2)
	// 		//t.Log("Decrypted: ", r2)
	// 	}
	// })


	t.Run("Check converting char '0' to int", func(t *testing.T){
		if r := tinycrypt.CharToInt('0'); r != 0 {
			t.Errorf("Wanted 0, got %d", r)
		}
	})

	t.Run("Check converting char '9' to int", func(t *testing.T){
		if r := tinycrypt.CharToInt('9'); r != 9 {
			t.Errorf("Wanted 9, got %d", r)
		}
	})

	t.Run("Check converting char 'A' to int", func(t *testing.T){
		if r := tinycrypt.CharToInt('A'); r != 10 {
			t.Errorf("Wanted 11, got %d", r)
		}
	})

	t.Run("Check converting char 'a' to int", func(t *testing.T){
		if r := tinycrypt.CharToInt('a'); r != 36{
			t.Errorf("Wanted 36, got %d", r)
		}
	})


	// test base case
	t.Run("Check hashing '00000000000000000000' to int", func(t *testing.T){
		if r := tinycrypt.MakeHash("00000000000000000000"); r != 0{
			t.Errorf("Wanted 0, got %d", r)
		}
	})

	// since A = 10 = b1010, the result should be a repetition of b001010 6 times
	t.Run("Check hashing 'A00A00A00A00A00A0000' to int", func(t *testing.T){
		if r := tinycrypt.MakeHash("A00A00A00A00A00A0000"); r != 10907853450{
			t.Errorf("Wanted 10907853450, got %d", r)
		}
	})

	// The last group of 6 bits should be 61 = 111101
	t.Run("Check hashing 'A00A00A00A00A00z0000' to int", func(t *testing.T){
		if r := tinycrypt.MakeHash("A00A00A00A00A00z0000"); r != 10907853501{
			t.Errorf("Wanted 10907853501, got %d", r)
		}
	})

	// The last group of 6 bits should still be 61 = 111101, since A + A = 10 + 10 = 20, and 20 / 41 = 0
	t.Run("Check hashing 'A00A00A00A00A00zAA00' to int", func(t *testing.T){
		if r := tinycrypt.MakeHash("A00A00A00A00A00zAA00"); r != 10907853501{
			t.Errorf("Wanted 10907853501, got %d", r)
		}
	})

	// The last group of 6 bits should now be 62 = 111110, since W + W = 32 + 32 = 64, and 64 / 41 = 1
	t.Run("Check hashing 'A00A00A00A00A00zWW00' to int", func(t *testing.T){
		if r := tinycrypt.MakeHash("A00A00A00A00A00zWW00"); r != 10907853502{
			t.Errorf("Wanted 109078535020, got %d", r)
		}
	})

	// The last group of 6 bits should now be 63 = 111111, since z + z = 61 + 61 = 122, and 122 / 41 = 2
	t.Run("Check hashing 'A00A00A00A00A00zzz00' to int", func(t *testing.T){
		if r := tinycrypt.MakeHash("A00A00A00A00A00zzz00"); r != 10907853503{
			t.Errorf("Wanted 10907853503, got %d", r)
		}
	})

	// test max value
	t.Run("Check hashing 'zzzzzzzzzzzzzzzzzz00' to int", func(t *testing.T){
		if r := tinycrypt.MakeHash("zzzzzzzzzzzzzzzzzz00"); r != 68719476735{
			t.Errorf("Wanted 68719476735, got %d", r)
		}
	})

	// test 3 words id generation
	t.Run("Check generation from 'zzzzzzzzzzzzzzzzzz00'", func(t *testing.T){
		r := tinycrypt.MakeHash("zzzzzzzzzzzzzzzzzz00")
		w := tinycrypt.GenerateWord36(r)
		t.Log(w)
	})

	// test 3 words id generation
	t.Run("Check generation from '00000000000000000000'", func(t *testing.T){
		r := tinycrypt.MakeHash("00000000000000000000")
		w := tinycrypt.GenerateWord36(r)
		t.Log(w)
	})



}

