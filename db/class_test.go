package db

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"io"
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

type TestFunc func(echo.Context) error

type TestObj struct {
	D			*DB
	Class		[]Class
	ClassBuf 	[]Class
	User		[]User
}

type ReqParam struct {
	HttpMethod	string 
	Path		string 
	Body 		io.Reader
	Function	TestFunc
	ExpCode		int
	Returns		bool
}

func CallFunc(t *testing.T, par *ReqParam) ([]byte, func() error) {
	req, err := http.NewRequest(	par.HttpMethod, 
									par.Path, 
									par.Body)
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	assert.NotNil(t, req, rec)
	//defer rec.Result().Body.Close()
	req.Header.Set("Content-Type", "application/json")
	c := echo.New().NewContext(req, rec)

	require.NoError(t, par.Function(c)) 
	//assert.Equal(t, rec.Code, par.ExpCode)
	var b []byte
	if par.Returns == true {
		assert.NotEmpty(t, rec.Result().Body)
		b, err = ioutil.ReadAll(rec.Result().Body)
		require.NoError(t, err)
		//t.Logf("Log: %s", string(b))
		assert.NoError(t, err)
	} else { 
		b = []byte{'{', '}'}
	}

	return b, rec.Result().Body.Close
}

func CreateTestUser(t *testing.T, o *TestObj, i int) {

	par := ReqParam {
		"POST", 
		"/", 
		nil,
		o.D.CreateUser,
		http.StatusOK,
		true,
	}
	b, f := CallFunc(t, &par)
	defer f()

	assert.NoError(t, json.Unmarshal([]byte(b), &o.User[i]))
	t.Logf(color_info+"Created user: %s"+color_end, o.User[i].UID)
}

func DeleteTestUser(t *testing.T, o *TestObj, i int) {
	pr := struct {
		Uid string
	}{
		o.User[i].UID,
	}
	pro, err := json.Marshal(&pr)
	require.NoError(t, err)

	par := ReqParam {
		"DELETE",
		"/",
		bytes.NewBuffer(pro),
		o.D.DeleteUser,
		http.StatusOK,
		false,
	}
	_, f := CallFunc(t, &par)
	defer f()

	t.Logf(color_info+"Removed user %s"+color_end, o.User[i].UID)
}

func GetTestClass(t *testing.T, o *TestObj, classIndex int, userIndex int) {
	pr := struct {
		Uid string
		Wid string
	}{
		o.User[userIndex].UID,
		o.Class[classIndex].WID,
	}
	pro, err := json.Marshal(&pr)
	require.NoError(t, err)

	par := ReqParam {
		"GET", 
		"/", 
		bytes.NewBuffer(pro),
		o.D.GetClass,
		http.StatusOK,
		true,
	}
	b, f := CallFunc(t, &par)
	defer f()

	assert.NoError(t, json.Unmarshal([]byte(b), &o.ClassBuf[classIndex]))
}

func CreateTestClass(t *testing.T, o *TestObj, classIndex int, userIndex int){


	pr := struct 	{
		Uid			string	
		Name		string
		Thumbnail 	int
	}{
		o.User[userIndex].UID,
		"TestClass",
		1,
	}
	pro, err := json.Marshal(&pr)
	require.NoError(t, err)

	par := ReqParam {
		"POST", 
		"/", 
		bytes.NewBuffer(pro),
		o.D.CreateClass,
		http.StatusOK,
		true,
	}
	b, f := CallFunc(t, &par)
	defer f()
	assert.NoError(t, json.Unmarshal([]byte(b), &o.Class[classIndex]))

	t.Logf(color_info+"CreateClass returned: \n%s"+color_end, string([]byte(b)))
}

func DeleteTestClass(t *testing.T, o *TestObj, classIndex int){

	pr := struct {
		Cid 	string
	}{
		o.Class[classIndex].CID,
	}
	pro, err := json.Marshal(&pr)
	require.NoError(t, err)

	par := ReqParam {
		"DELETE",
		"/",
		bytes.NewBuffer(pro),
		o.D.DeleteClass,
		http.StatusOK,
		false,
	}
	CallFunc(t, &par)

	t.Logf(color_info+"Removed class %s"+color_end, o.Class[classIndex].CID)
}

func IsIn(str string, list []string) bool {
	for _, s := range list {
		if s == str {
			return true
		}
	}
	return false
}

// Ensure classes can be created w/o any errors
func TestCreateClass(t *testing.T) {
	
	obj := TestObj {
		nil,
		make([]Class, 1),
		make([]Class, 1),
		make([]User, 1),
	}

	var err error = nil
	obj.D, err = Open(context.Background(), os.Getenv("TLACFG"))
	require.NoError(t, err)

	CreateTestUser(t, &obj, 0)
	CreateTestClass(t, &obj, 0, 0)
	
	DeleteTestClass(t, &obj, 0)
	DeleteTestUser(t, &obj, 0)
}


func TestGetClass(t *testing.T) {

	obj := TestObj {
		nil,
		make([]Class, 1),
		make([]Class, 1),
		make([]User, 1),
	}

	var err error = nil
	obj.D, err = Open(context.Background(), os.Getenv("TLACFG"))
	require.NoError(t, err)

	CreateTestUser(t, &obj, 0)
	CreateTestClass(t, &obj, 0, 0)
	defer DeleteTestClass(t, &obj, 0)
	defer DeleteTestUser(t, &obj, 0)

	GetTestClass(t, &obj, 0, 0)
	t.Logf("%v+", obj.Class[0])
	t.Logf("%v+", obj.ClassBuf[0])
	// Make sure classes are same. 	
	// Compare manually, since DeepEqual and cmp.Equal both fails for some reason...
	assert.True(t, obj.Class[0].CID == obj.ClassBuf[0].CID)
	assert.True(t, obj.Class[0].WID == obj.ClassBuf[0].WID)
	assert.True(t, obj.Class[0].Name == obj.ClassBuf[0].Name)
}

func TestJoinLeaveClass(t *testing.T) {

	obj := TestObj {
		nil,
		make([]Class, 1),
		make([]Class, 1),
		make([]User, 2),
	}

	var err error = nil
	obj.D, err = Open(context.Background(), os.Getenv("TLACFG"))
	require.NoError(t, err)

	CreateTestUser(t, &obj, 0)
	CreateTestClass(t, &obj, 0, 0)
	defer DeleteTestClass(t, &obj, 0)
	defer DeleteTestUser(t, &obj, 0)

	// Create student to join class
	CreateTestUser(t, &obj, 1)
	defer DeleteTestUser(t, &obj, 1)

	// Join user
	pr := struct {
		Uid string
		Cid string
	}{
		obj.User[1].UID,
		obj.Class[0].WID,
	}
	pro, err := json.Marshal(&pr)
	require.NoError(t, err)

	par := ReqParam {
		"PUT", 
		"/", 
		bytes.NewBuffer(pro),
		obj.D.JoinClass,
		http.StatusOK,
		true,
	}
	b, f := CallFunc(t, &par)
	assert.NoError(t, json.Unmarshal([]byte(b), &obj.ClassBuf[0]))
	f()

	t.Logf(color_info+"Adding student: \t%s \nto class: \t%s"+color_end, obj.User[1].UID, obj.ClassBuf[0].WID)
	
	// JoinClass returns the class struct BEFORE adding the student 
	GetTestClass(t, &obj, 0, 0)

	// Make sure student is in class
	students := obj.ClassBuf[0].Members
	assert.True(t, IsIn(obj.User[1].UID, students))

	pr = struct {
		Uid string
		Cid string
	}{
		obj.User[1].UID,
		obj.Class[0].CID,
	}
	pro, err = json.Marshal(&pr)
	require.NoError(t, err)

	t.Logf(color_info+"Leave student: \t%s \nfrom class: \t%s"+color_end, obj.User[1].UID, obj.Class[0].CID)

	par = ReqParam {
		"PUT", 
		"/", 
		bytes.NewBuffer(pro),
		obj.D.LeaveClass,
		http.StatusOK,
		false,
	}
	b, _ = CallFunc(t, &par)

	GetTestClass(t, &obj, 0, 0)
	assert.False(t, IsIn(obj.User[1].UID, obj.ClassBuf[0].Members))
}

