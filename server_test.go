package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/uclaacm/teach-la-go-backend/db"
)

// PORT defines where we serve the backend.
const PORT = ":8081"

//structure to store
type Response struct {
	UserData db.User
	Programs []string
}

//Runs series of test to test functionality of database
func TestRigDB(t *testing.T) {

	var (
		d           *db.DB // stores instance of connection with database
		err         error
		res         Response // structure to store response fron database
		res_student Response // structure to store response fron database
		class_id    string
	)

	t.Logf("Testing initialization of database...")

	// Test opening connection with database
	t.Run("Open connection with database", func(t *testing.T) {
		if d, err = db.OpenFromEnv(context.Background()); err != nil {
			t.Fatal("failed to open DB client")
		}
	})

	t.Logf("Testing creating new user...")

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

	t.Run("Delete program", func(t *testing.T) {

		body := struct {
			UID string `json:"uid"`
			PID string `json:"pid"`
		}{
			UID: res.UserData.UID,
			PID: res.UserData.Programs[0],
		}

		j, err := json.Marshal(&body)
		if err != nil {
			t.Fatal("Failed to marshal request body")
		}

		req, err := http.NewRequest("DELETE", "/program/delete", bytes.NewBuffer(j))

		if err != nil {
			t.Fatal("Failed to create http request")
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(d.HandleDeleteProgram)

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Fatal("Delete program failed")
		}

	})

	t.Logf("Testing creating program for user...")

	t.Run("Create program", func(t *testing.T) {

		t.Logf("Building query")
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

		req, err := http.NewRequest("POST", "/program/create", bytes.NewBuffer(pro))
		req.Header.Set("Content-Type", "application/json")

		if err != nil {
			t.Fatal("Failed to test create program")
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(d.HandleInitializeProgram)

		t.Logf("Making call...")
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Fatal("Create program failed")
		}

	})

	t.Logf("Testing getting user...")

	t.Run("Get user", func(t *testing.T) {

		t.Logf("Building query")

		req, err := http.NewRequest("GET", "/user/get", nil)

		if err != nil {
			t.Fatal("Failed to create http request")
		}

		// build query
		p := req.URL.Query()

		//p.Add("userId", res.UserData.UID)
		//p.Add("includePrograms", res.UserData.Programs[0])
		t.Log(res.UserData.UID)
		p.Add("uid", res.UserData.UID)
		p.Add("programs", "true")
		req.URL.RawQuery = p.Encode()

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(d.HandleGetUser)

		// struct to recieve response
		resp := struct {
			UserData *db.User     `json:"userData"`
			Programs []db.Program `json:"programs"`
		}{}

		handler.ServeHTTP(rr, req)
		t.Log(rr.Body)
		//get raw data returned
		//unmarshall json
		j, err := ioutil.ReadAll(rr.Result().Body)
		if err != nil {
			t.Fatal("Failed to read response")
		}

		json.Unmarshal([]byte(j), &resp)

		//TODO check if correct programs are made
		//t.Logf(resp.Programs[0].Name)

		if status := rr.Code; status != http.StatusOK {
			t.Fatal("Get user failed")
		}

	})

	// Test creating a class from a user
	t.Run("Create Class", func(t *testing.T) {

		// create JSON for a new program
		pr := struct {
			Uid       string
			Name      string
			Thumbnail int
		}{
			res.UserData.UID,
			"TestClass",
			1,
		}

		pro, err := json.Marshal(&pr)

		if err != nil {
			t.Fatal("Failed to create JSON")
		}

		//fmt.Printf("%s", pro)

		req, err := http.NewRequest("POST", "/class/create", bytes.NewBuffer(pro))
		req.Header.Set("Content-Type", "application/json")

		if err != nil {
			t.Fatal("Failed to test create program")
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(d.HandleCreateClass)

		handler.ServeHTTP(rr, req)

		var class db.Class

		t.Log(rr.Body)

		j, err := ioutil.ReadAll(rr.Result().Body)
		if err != nil {
			t.Fatal("Failed to read response")
		}

		json.Unmarshal([]byte(j), &class)
		class_id = class.WID

		if status := rr.Code; status != http.StatusOK {
			t.Fatal("Create program failed")
		}

	})

	// Create another student to join the class
	t.Run("Create new student", func(t *testing.T) {
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

		defer rr.Result().Body.Close()

		j, err := ioutil.ReadAll(rr.Result().Body)
		if err != nil {
			t.Fatal("Failed to read response")
		}

		json.Unmarshal([]byte(j), &res_student)
	})

	// Test adding a user to class
	t.Run("Join Class", func(t *testing.T) {

		// create JSON for a new program
		pr := struct {
			Uid string
			Cid string
		}{
			res.UserData.UID,
			class_id,
		}

		pro, err := json.Marshal(&pr)

		if err != nil {
			t.Fatal("Failed to create JSON")
		}

		//fmt.Printf("%s", pro)

		req, err := http.NewRequest("POST", "/class/join", bytes.NewBuffer(pro))
		req.Header.Set("Content-Type", "application/json")

		if err != nil {
			t.Fatal("Failed to test create program")
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(d.HandleJoinClass)

		handler.ServeHTTP(rr, req)
		t.Log(rr.Body)
		if status := rr.Code; status != http.StatusOK {
			t.Fatal("Create program failed")
		}

	})

}
