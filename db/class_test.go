package db

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"bytes"

	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

)


/* Variables to store data persistant across tests*/
var (
	class Class
	class_ret Class 
	user_ClassOwner User
	user User
)

func TestCreateClass(t *testing.T) {
	d, err := Open(context.Background(), os.Getenv("TLACFG"))
	require.NoError(t, err)

	// Create a new user to host the class
	t.Run("Create a new user to host class", func(t *testing.T) {
	
		req, err := http.NewRequest("POST", "/", nil)
		if err != nil {
			t.Fatal("Failed to create http request")
		}
		
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.CreateUser(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.NotEmpty(t, rec.Result().Body)
			defer rec.Result().Body.Close()
			/* store the user */
			j, err := ioutil.ReadAll(rec.Result().Body)
			if err != nil {
				t.Fatal("Failed to read response")
			}
			json.Unmarshal([]byte(j), &user_ClassOwner)
		}
	
	})

	// Test creating a class from the user
	t.Run("Create Class", func(t *testing.T){
		// create JSON for a new program 
		pr := struct {
			Uid 		string
			Name 		string
			Thumbnail	int
		}{
			user_ClassOwner.UID,
			"TestClass",
			1,
		}

		pro, err := json.Marshal(&pr) 
		if err != nil {
			t.Fatal("Failed to create JSON")
		}

		req, err := http.NewRequest("POST", "/", bytes.NewBuffer(pro))
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			t.Fatal("Failed to test create program")
		}
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.CreateClass(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotEmpty(t, rec.Result().Body)
			/* store the class */
			j, err := ioutil.ReadAll(rec.Result().Body)
			if err != nil {
				t.Fatal("Failed to read response")
			}
			json.Unmarshal([]byte(j), &class)
			t.Logf(string([]byte(j)))
			defer rec.Result().Body.Close()
		}
	})
}

func TestJoinClass(t *testing.T) {
	d, err := Open(context.Background(), os.Getenv("TLACFG"))
	require.NoError(t, err)

	// Create a new user to join the class
	t.Run("Create a new user to join class", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/", nil)
		if err != nil {
			t.Fatal("Failed to create http request")
		}
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.CreateUser(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.NotEmpty(t, rec.Result().Body)
			defer rec.Result().Body.Close()
			/* store the user */
			j, err := ioutil.ReadAll(rec.Result().Body)
			if err != nil {
				t.Fatal("Failed to read response")
			}
			json.Unmarshal([]byte(j), &user)
		}
	
	})

	// Test adding a user to class
	t.Run("Add student to class", func(t *testing.T){

		// create JSON to request adding user to class
		pr := struct {
			Uid 		string
			Cid			string
		}{
			user.UID,
			class.WID,
		}

		pro, err := json.Marshal(&pr) 
		if err != nil {
			t.Fatal("Failed to create JSON")
		}
		t.Logf(string(pro))
		req, err := http.NewRequest("PUT", "/", bytes.NewBuffer(pro))
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			t.Fatal("Failed to create http request")
		}
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.JoinClass(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotEmpty(t, rec.Result().Body)
			defer rec.Result().Body.Close()
		}

	})

}

func TestGetClass(t *testing.T) {
	d, err := Open(context.Background(), os.Getenv("TLACFG"))
	require.NoError(t, err)

	t.Run("Send request to get class", func(t *testing.T) {
		// create JSON to get class 
		pr := struct {
			Uid 		string
			Wid 		string
		}{
			user_ClassOwner.UID,
			class.WID,
		}
		pro, err := json.Marshal(&pr) 
		if err != nil {
			t.Fatal("Failed to create JSON")
		}
		t.Logf(string(pro))

		req, err := http.NewRequest("GET", "/", bytes.NewBuffer(pro))
		if err != nil {
			t.Fatal("Failed to create http request")
		}
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.GetClass(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotEmpty(t, rec.Result().Body)
			defer rec.Result().Body.Close()
			/* store the class */
			j, err := ioutil.ReadAll(rec.Result().Body)
			if err != nil {
				t.Fatal("Failed to read response")
			}
			json.Unmarshal([]byte(j), &class_ret)
		}
	})

	t.Run("Check contents of returned class", func(t *testing.T){

		// TODO make sure class == class_ret

	})

}

func TestLeaveClass(t *testing.T) {
	d, err := Open(context.Background(), os.Getenv("TLACFG"))
	require.NoError(t, err)

	t.Run("Remove a student from a class", func(t *testing.T) {
		// create JSON to get class 
		pr := struct {
			Uid 		string
			Cid 		string
		}{
			user.UID,
			class.CID,
		}
		pro, err := json.Marshal(&pr) 
		if err != nil {
			t.Fatal("Failed to create JSON")
		}

		req, err := http.NewRequest("PUT", "/", bytes.NewBuffer(pro))
		if err != nil {
			t.Fatal("Failed to create http request")
		}
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.LeaveClass(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotEmpty(t, rec.Result().Body)
			defer rec.Result().Body.Close()
		}
	})

	t.Run("Get class...", func(t *testing.T) {
		// create JSON to get class 
		pr := struct {
			Uid 		string
			Wid 		string
		}{
			user_ClassOwner.UID,
			class.WID,
		}
		pro, err := json.Marshal(&pr) 
		if err != nil {
			t.Fatal("Failed to create JSON")
		}

		req, err := http.NewRequest("GET", "/", bytes.NewBuffer(pro))
		if err != nil {
			t.Fatal("Failed to create http request")
		}
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.GetClass(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotEmpty(t, rec.Result().Body)
			defer rec.Result().Body.Close()
			/* store the class */
			j, err := ioutil.ReadAll(rec.Result().Body)
			if err != nil {
				t.Fatal("Failed to read response")
			}
			json.Unmarshal([]byte(j), &class_ret)
		}
	})

	t.Run("... and check its contents", func(t *testing.T){

		// TODO fail if the deleted student is still in class

	})

}

func TestClassCleanup(t *testing.T) {
	d, err := Open(context.Background(), os.Getenv("TLACFG"))
	require.NoError(t, err)

	t.Run("Remove the test user", func(t *testing.T) {
		pr := struct {
			Uid 		string
		}{
			user.UID,
		}
		pro, err := json.Marshal(&pr) 
		if err != nil {
			t.Fatal("Failed to create JSON")
		}
		t.Logf(user.UID)
		req, err := http.NewRequest("DELETE", "/", bytes.NewBuffer(pro))
		if err != nil {
			t.Fatal("Failed to create http request")
		}
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.DeleteUser(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotEmpty(t, rec.Result().Body)
			t.Logf(string(rec.Code))
			//defer rec.Result().Body.Close()
		}
	})

	t.Run("Remove the class", func(t *testing.T) {
		pr := struct {
			Cid 		string
		}{
			class.CID,
		}
		pro, err := json.Marshal(&pr) 
		if err != nil {
			t.Fatal("Failed to create JSON")
		}

		req, err := http.NewRequest("DELETE", "/", bytes.NewBuffer(pro))
		if err != nil {
			t.Fatal("Failed to create http request")
		}
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.DeleteClass(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			defer rec.Result().Body.Close()
		}
	})

	t.Run("Remove the class owner", func(t *testing.T) {
		pr := struct {
			Uid 		string
		}{
			user_ClassOwner.UID,
		}
		pro, err := json.Marshal(&pr) 
		if err != nil {
			t.Fatal("Failed to create JSON")
		}

		req, err := http.NewRequest("DELETE", "/", bytes.NewBuffer(pro))
		if err != nil {
			t.Fatal("Failed to create http request")
		}
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.DeleteUser(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			defer rec.Result().Body.Close()
		}
	})
}

