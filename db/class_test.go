package db

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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
	colorInfo = "\033[32m"
	colorEnd  = "\033[0m"
)

type TestFunc func(echo.Context) error

type TestObj struct {
	D        *DB
	Class    []Class
	ClassBuf []Class
	User     []User
}

type ReqParam struct {
	HTTPMethod string
	Path       string
	Body       io.Reader
	Function   TestFunc // function to close response body
	ExpCode    int      // expected return code.
	Returns    bool     // specify if this call returns a body or not
}

func CallFunc(t *testing.T, par *ReqParam) ([]byte, func() error) {
	req := httptest.NewRequest(par.HTTPMethod,
		par.Path,
		par.Body)
	rec := httptest.NewRecorder()
	assert.NotNil(t, req, rec)
	req.Header.Set("Content-Type", "application/json")
	c := echo.New().NewContext(req, rec)

	require.NoError(t, par.Function(c))
	assert.Equal(t, par.ExpCode, rec.Code)
	var b []byte
	if par.Returns == true {
		assert.NotEmpty(t, rec.Result().Body)
		var err error
		b, err = ioutil.ReadAll(rec.Result().Body)
		require.NoError(t, err)
		assert.NoError(t, err)
	} else {
		b = []byte{'{', '}'}
	}

	return b, rec.Result().Body.Close
}

func CreateTestUser(t *testing.T, o *TestObj, i int) {

	par := ReqParam{
		"POST",
		"/",
		nil,
		o.D.CreateUser,
		http.StatusCreated,
		true,
	}
	b, close := CallFunc(t, &par)
	defer assert.NoError(t, close())

	assert.NoError(t, json.Unmarshal([]byte(b), &o.User[i]))
	t.Logf(colorInfo+"Created user: %s"+colorEnd, o.User[i].UID)
}

// func DeleteTestUser(t *testing.T, o *TestObj, i int) {
// 	pr := struct {
// 		UID string
// 	}{
// 		o.User[i].UID,
// 	}
// 	pro, err := json.Marshal(&pr)
// 	require.NoError(t, err)

// 	par := ReqParam{
// 		"DELETE",
// 		"/",
// 		bytes.NewBuffer(pro),
// 		o.D.DeleteUser,
// 		http.StatusOK,
// 		false,
// 	}
// 	_, close := CallFunc(t, &par)
// 	defer assert.NoError(t, close())

// 	t.Logf(colorInfo+"Removed user %s"+colorEnd, o.User[i].UID)
// }

func CreateTestClass(t *testing.T, o *TestObj, classIndex int, userIndex int) {
	pr := struct {
		UID       string
		Name      string
		Thumbnail int
	}{
		o.User[userIndex].UID,
		"TestClass",
		1,
	}
	pro, err := json.Marshal(&pr)
	require.NoError(t, err)

	par := ReqParam{
		"POST",
		"/",
		bytes.NewBuffer(pro),
		o.D.CreateClass,
		http.StatusOK,
		true,
	}
	b, close := CallFunc(t, &par)
	defer assert.NoError(t, close())
	assert.NoError(t, json.Unmarshal([]byte(b), &o.Class[classIndex]))

	t.Logf(colorInfo+"CreateClass returned: \n%s"+colorEnd, string([]byte(b)))
}

// func DeleteTestClass(t *testing.T, o *TestObj, classIndex int) {
// 	pr := struct {
// 		Cid string
// 	}{
// 		o.Class[classIndex].CID,
// 	}
// 	pro, err := json.Marshal(&pr)
// 	require.NoError(t, err)

// 	par := ReqParam{
// 		"DELETE",
// 		"/",
// 		bytes.NewBuffer(pro),
// 		o.D.DeleteClass,
// 		http.StatusOK,
// 		false,
// 	}
// 	_, close := CallFunc(t, &par)
// 	defer assert.NoError(t, close())

// 	t.Logf(colorInfo+"Removed class %s"+colorEnd, o.Class[classIndex].CID)
// }

// func IsIn(str string, list []string) bool {
// 	for _, s := range list {
// 		if s == str {
// 			return true
// 		}
// 	}
// 	return false
// }

// Ensure classes can be created w/o any errors
func TestCreateClass(t *testing.T) {

	obj := TestObj{
		nil,
		make([]Class, 1),
		make([]Class, 1),
		make([]User, 1),
	}

	ptr, err := Open(context.Background(), os.Getenv("TLACFG"))
	obj.D = ptr
	require.NoError(t, err)

	CreateTestUser(t, &obj, 0)
	CreateTestClass(t, &obj, 0, 0)

	// DeleteTestClass(t, &obj, 0)
	// DeleteTestUser(t, &obj, 0)
}
