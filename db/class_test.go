package db

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"

	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	color_info = "\033[32m"
	color_end  = "\033[0m"
)

/* Variables to store data persistant across tests*/
var (
	class          Class
	classRet       Class
	userClassOwner User
	user           User
)

func TestCreateClass(t *testing.T) {
	d, err := Open(context.Background(), os.Getenv("TLACFG"))
	require.NoError(t, err)

	t.Run("Create a new user to host class", func(t *testing.T) {

		req, err := http.NewRequest("POST", "/", nil)
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)

		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.CreateUser(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.NotEmpty(t, rec.Result().Body)
			j, err := ioutil.ReadAll(rec.Result().Body)
			defer rec.Result().Body.Close()

			require.NoError(t, err)
			assert.NoError(t, json.Unmarshal([]byte(j), &userClassOwner))

			t.Logf(color_info+"Created class owner user: %s"+color_end, userClassOwner.UID)
		}
	})


	t.Run("Init shards", func(t *testing.T) {
		err := d.InitShards(context.Background(), classesAliasPath)
		assert.Nil(t, err)
	})

	t.Run("Create Class", func(t *testing.T) {
		pr := struct {
			Uid       string
			Name      string
			Thumbnail int
		}{
			userClassOwner.UID,
			"TestClass",
			1,
		}
		pro, err := json.Marshal(&pr)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/", bytes.NewBuffer(pro))
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)

		req.Header.Set("Content-Type", "application/json")
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.CreateClass(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotEmpty(t, rec.Result().Body)
			j, err := ioutil.ReadAll(rec.Result().Body)
			defer rec.Result().Body.Close()

			require.NoError(t, err)
			assert.NoError(t, json.Unmarshal([]byte(j), &class))

			t.Logf(color_info+"CreateClass returned: \n%s"+color_end, string([]byte(j)))
		}
	})
}

func TestJoinClass(t *testing.T) {
	d, err := Open(context.Background(), os.Getenv("TLACFG"))
	require.NoError(t, err)

	t.Run("Create a new user to join class", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/", nil)
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)

		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.CreateUser(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.NotEmpty(t, rec.Result().Body)
			j, err := ioutil.ReadAll(rec.Result().Body)
			defer rec.Result().Body.Close()

			require.NoError(t, err)
			assert.NoError(t, json.Unmarshal([]byte(j), &user))

			t.Logf(color_info+"JoinClass returned: \n%s"+color_end, string([]byte(j)))
		}
	})

	t.Run("Add student to class", func(t *testing.T) {
		pr := struct {
			Uid string
			Cid string
		}{
			user.UID,
			class.WID,
		}
		pro, err := json.Marshal(&pr)
		require.NoError(t, err)

		t.Logf(color_info+"Adding student: \t%s \nto class: \t%s"+color_end, user.UID, class.WID)

		req, err := http.NewRequest("PUT", "/", bytes.NewBuffer(pro))
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)

		req.Header.Set("Content-Type", "application/json")
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
		pr := struct {
			Uid string
			Wid string
		}{
			userClassOwner.UID,
			class.WID,
		}
		pro, err := json.Marshal(&pr)
		require.NoError(t, err)

		t.Logf(color_info+"Get class: %s"+color_end, class.WID)

		req, err := http.NewRequest("GET", "/", bytes.NewBuffer(pro))
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)

		req.Header.Set("Content-Type", "application/json")
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.GetClass(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotEmpty(t, rec.Result().Body)
			j, err := ioutil.ReadAll(rec.Result().Body)
			defer rec.Result().Body.Close()

			require.NoError(t, err)
			assert.NoError(t, json.Unmarshal([]byte(j), &classRet))
		}
	})

	t.Run("Check contents of returned class", func(t *testing.T) {
		assert.Equal(t, class.CID, classRet.CID)
		assert.Equal(t, class.WID, classRet.WID)
	})
}

func TestLeaveClass(t *testing.T) {
	d, err := Open(context.Background(), os.Getenv("TLACFG"))
	require.NoError(t, err)

	t.Run("Remove a student from a class", func(t *testing.T) {
		pr := struct {
			Uid string
			Cid string
		}{
			user.UID,
			class.CID,
		}
		pro, err := json.Marshal(&pr)
		require.NoError(t, err)

		t.Logf(color_info+"Leave student: \t%s \nfrom class: \t%s"+color_end, user.UID, class.WID)

		req, err := http.NewRequest("PUT", "/", bytes.NewBuffer(pro))
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)

		req.Header.Set("Content-Type", "application/json")
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.LeaveClass(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotEmpty(t, rec.Result().Body)
			defer rec.Result().Body.Close()
		}
	})

	t.Run("Get class...", func(t *testing.T) {
		pr := struct {
			Uid string
			Wid string
		}{
			userClassOwner.UID,
			class.WID,
		}
		pro, err := json.Marshal(&pr)
		require.NoError(t, err)

		req, err := http.NewRequest("GET", "/", bytes.NewBuffer(pro))
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)

		req.Header.Set("Content-Type", "application/json")
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.GetClass(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotEmpty(t, rec.Result().Body)
			j, err := ioutil.ReadAll(rec.Result().Body)
			defer rec.Result().Body.Close()

			require.NoError(t, err)
			assert.NoError(t, json.Unmarshal([]byte(j), &classRet))
			assert.Equal(t, class.CID, classRet.CID)
		}
	})

	t.Run("... and check its contents", func(t *testing.T) {
		assert.Empty(t, classRet.Members)
		// TODO: more testing
	})
}

func TestClassCleanup(t *testing.T) {
	d, err := Open(context.Background(), os.Getenv("TLACFG"))
	require.NoError(t, err)

	t.Run("Remove the test user", func(t *testing.T) {
		pr := struct {
			Uid string
		}{
			user.UID,
		}
		pro, err := json.Marshal(&pr)
		require.NoError(t, err)

		t.Logf(color_info+"Removing user %s"+color_end, user.UID)

		req, err := http.NewRequest("DELETE", "/", bytes.NewBuffer(pro))
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)

		req.Header.Set("Content-Type", "application/json")
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.DeleteUser(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotEmpty(t, rec.Result().Body)
			defer rec.Result().Body.Close()
		}
	})

	t.Run("Remove the class", func(t *testing.T) {
		pr := struct {
			Cid string
		}{
			class.CID,
		}
		pro, err := json.Marshal(&pr)
		require.NoError(t, err)

		t.Logf(color_info+"Removing class %s"+color_end, class.CID)

		req, err := http.NewRequest("DELETE", "/", bytes.NewBuffer(pro))
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)

		req.Header.Set("Content-Type", "application/json")
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.DeleteClass(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			defer rec.Result().Body.Close()
		}
	})

	t.Run("Remove the class owner", func(t *testing.T) {
		pr := struct {
			Uid string
		}{
			userClassOwner.UID,
		}
		pro, err := json.Marshal(&pr)
		require.NoError(t, err)

		t.Logf(color_info+"Removing user %s"+color_end, userClassOwner.UID)

		req, err := http.NewRequest("DELETE", "/", bytes.NewBuffer(pro))
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)

		req.Header.Set("Content-Type", "application/json")
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.DeleteUser(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			defer rec.Result().Body.Close()
		}
	})
}
