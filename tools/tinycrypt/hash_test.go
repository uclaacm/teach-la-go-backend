package tinycrypt_test

import (
	"testing"
	"../tinycrypt"
)

func TestEncrypt(t *testing.T) {

	var (
		encrypt_hash tinycrypt.Encrypter
		res int
	)

	res = 0
	

	//set key 
	encrypt_hash.InitializeEncrypter(0x33684192D)

	t.Logf("Encrypting 12345...")
	
	t.Run("Encrypt 12345", func(t *testing.T){
		t.Log("Input: ", res)
		res = encrypt_hash.Encrypt36(res)
		t.Log("Result: ", res)
	})
	res = 1
	t.Run("Encrypt 12345", func(t *testing.T){
		t.Log("Input: ", res)
		res = encrypt_hash.Encrypt36(res)
		t.Log("Result: ", res)
	})
	res = 2
	t.Run("Encrypt 12345", func(t *testing.T){
		t.Log("Input: ", res)
		res = encrypt_hash.Encrypt36(res)
		t.Log("Result: ", res)
	})

	var r2 uint8 
	r2 = 0
	
	t.Run("Encrypt loop", func(t *testing.T){
		for i := 0; i < 256; i++{
			r2 = uint8(i)
			//t.Log("Input: ", r2)
			r2 = encrypt_hash.Encrypt8(r2)
			t.Log("Result: ", r2)
		}
	})
	

	/*
	t.Run("Decrypt 12345", func(t *testing.T){
		t.Log("Input: ", res)
		res = encrypt_hash.Encrypt(res)
		t.Log("Result: ", res)
	})*/
	
	
	
	

}

