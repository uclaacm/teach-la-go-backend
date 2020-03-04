package main_test

import (
	"context"
	"testing"
	"strings"

	"./db"
	tinycrypt "github.com/uclaacm/teach-la-go-backend-tinycrypt" 
)

//Runs series of test to test functionality of database
func TestAliasDB(t *testing.T) {

	var (
		d   *db.DB 		// stores instance of connection with database
		err error
	)
	
	t.Logf("Testing initialization of database...")

	// Test opening connection with database
	t.Run("Open connection with database", func(t *testing.T) {
		if d, err = db.OpenFromEnv(context.Background()); err != nil {
			t.Fatal("failed to open DB client")
		}
	})

		
	//Test creating a class from a user
	t.Run("Create Shards", func(t *testing.T){
		err := d.InitShards(context.Background(), "classes_alias")
		if err != nil {
			t.Fatal("init failed")
		}
	})

	// Test creating a class from a user
	t.Run("Get ID", func(t *testing.T){
		for i := 0; i < 32; i++ {
			u, err := d.GetID(context.Background(), "classes_alias")
			if err != nil {
				t.Fatal(err)
			}
			
			t.Logf("u: %d\n",u)
			w := tinycrypt.GenerateWord24(uint64(u))
			wid := strings.Join(w, ",") 
			t.Logf("u: %s\n======",wid)

		}
	})

	
}
