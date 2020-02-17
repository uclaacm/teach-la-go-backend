package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"./lib"
)

// PORT defines where we serve the backend.
const PORT = ":8081"

//structure to store
type Response struct {
	UserData lib.User
	Programs []string
}

//Runs series of test to test functionality of database
func TestRigDB(t *testing.T) {

	var (
		d   *lib.DB // stores instance of connection with database
		err error
		res Response // structure to store response fron database
	)

	t.Logf("Testing initialization of database...")

	// Test opening connection with database
	t.Run("Open connection with database", func(t *testing.T) {

		if d, err = lib.OpenFromEnv(context.Background()); err != nil {
			t.Fatal("failed to open DB client")
		}
	})

	t.Logf("Testing creating new user...")

	// Test creating a new user
	t.Run("Create new user", func(t *testing.T) {

		req, err := http.NewRequest("POST", "/user/create", nil)

		if err != nil {
			t.Fatal("Failed to create http request")
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(d.HandleInitializeUser)

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Fatal("Failed to create user")
		}

		// if creating user succeeded, record the response to use it in the next test
		defer rr.Result().Body.Close()

		t.Logf("Create user successful")
		t.Logf("Reading response")

		j, err := ioutil.ReadAll(rr.Result().Body)
		if err != nil {
			t.Fatal("Failed to read response")
		}

		json.Unmarshal([]byte(j), &res)

	})

	t.Logf("Testing deletion of program from user...")

	// Test deleting a program from a user
	t.Run("Delete program", func(t *testing.T) {

		req, err := http.NewRequest("DELETE", "/program/delete", nil)

		if err != nil {
			t.Fatal("Failed to create http request")
		}

		// build query
		p := req.URL.Query()

		//fmt.Printf("UID: %s\n", res.UserData.UID)
		//fmt.Printf("Program: %s\n", res.UserData.Programs[0])

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

	t.Logf("Testing creating program for user...")

	// Test creating a program from a user
	t.Run("Create program", func(t *testing.T) {

		// create JSON for a new program
		pr := struct {
			Code        string
			DateCreated string
			Language    string
			Name        string
			Thumbnail   int
			Uid         string
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

		//fmt.Printf("%s", pro)

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
