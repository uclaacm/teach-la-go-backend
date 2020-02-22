package tinycrypt_test

import (
	"testing"
	"../tinycrypt"
)

func TestEncrypt(t *testing.T) {

	var en tinycrypt.Encrypter
	var de tinycrypt.Encrypter

	keys := []uint64 {
		0x33684192D,
		0x28DAB6A5A,
	}

	keys_rev := []uint64 {
		0x28DAB6A5A,
		0x33684192D,
	}


	en.InitializeEncrypter(keys)
	de.InitializeEncrypter(keys_rev)

	var res uint64
	res = 0	

	t.Logf("Encrypting...")
	
	for i := 1; i < 256; i++{
		res = uint64(i)
		t.Run("Encrypt", func(t *testing.T){
		
			res = en.Encrypt36(res)
			t.Log("Encrypted: ", res)
		
			res = de.Encrypt36(res)
			t.Log("Decrypted: ", res)
		})
		
	}

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

}

