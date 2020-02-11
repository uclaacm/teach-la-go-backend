package tester

import(
	"context"
	"bytes"
	//"os"
	"testing"
	"net/http"
	"net/http/httptest"
	"encoding/json"
	"io/ioutil"
	"fmt"

	"../lib"
)

// PORT defines where we serve the backend.
const PORT = ":8081"

type Response struct{
	UserData lib.User 
	Programs []string
}
//Runs series of test to test functionality of database
func TestRigDB(t *testing.T) {

	// set up context for main routine.
	//mainContext := context.Background()

	// acquire DB client.
	// fails early if we cannot acquire one.
	var (
		d   *lib.DB
		err error
	)

	var res Response

	// Test opening connection with database
	t.Run("Open connection with database", func(t *testing.T){

		if d, err = lib.OpenFromEnv(context.Background()); err != nil {
			t.Fatal("failed to open DB client")
		}	
	})
	//defer d.Close()

	

	// Test creating a new user
	t.Run("Create new user", func(t *testing.T){

		req, err := http.NewRequest("POST", "/user/create", nil)

		if err != nil {
			t.Fatal("Failed to create http request")
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(d.HandleInitializeUser)

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Fatal("Failed to create users")
		}

		defer rr.Result().Body.Close()

		j, err := ioutil.ReadAll(rr.Result().Body)
		if err != nil {
		 	t.Fatal("Failed to create users")
		}

		json.Unmarshal([]byte(j), &res)

	})

	

	// Test deleting a program from a user
	t.Run("Delete program", func(t *testing.T){

		req, err := http.NewRequest("DELETE", "/program/delete", nil)

		if err != nil {
			t.Fatal("Failed to test create user")
		}

		p := req.URL.Query()

		fmt.Printf("UID: %s\n", res.UserData.UID)
		fmt.Printf("Program: %s\n", res.UserData.Programs[0])
		
		p.Add("userId", res.UserData.UID)
		p.Add("programId", res.UserData.Programs[0])
		req.URL.RawQuery = p.Encode()

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(d.HandleDeleteProgram)

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Fatal("Delete program failed")
		}

	})

	// Test creating a program from a user
	t.Run("Create program", func(t *testing.T){

		pr := struct {
			Code 		string 
			DateCreated string
			Language 	string 
			Name 		string 
			Thumbnail 	int
			Uid 		string
		}{
			"print(my cool code)\n",
			"2020-02-07T08:41:00Z",
			"python",
			"Neat Program",
			12,
			res.UserData.UID,
		}

		pro, err := json.Marshal(&pr) 

		if err != nil {
			t.Fatal("Failed to create JSON")
		}

		fmt.Printf("%s", pro)

		req, err := http.NewRequest("POST", "/program/create", bytes.NewBuffer(pro))
		req.Header.Set("Content-Type", "application/json")

		if err != nil {
			t.Fatal("Failed to test create program")
		}
		
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(d.HandleInitializeProgram)

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Fatal("Create program failed")
		}

	})



}

