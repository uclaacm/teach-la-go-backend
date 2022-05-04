package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/uclaacm/teach-la-go-backend/db"
	"github.com/uclaacm/teach-la-go-backend/handler"
)

func TestGetProgram(t *testing.T) {
	t.Run("MissingPID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)
		d := db.OpenMock()

		if assert.NoError(t,
			handler.GetProgram(
				&db.DBContext{
					Context: c,
					TLADB:   d,
				}),
		) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})

	t.Run("BadPID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?pid=fakePID", nil)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)
		d := db.OpenMock()

		if assert.NoError(
			t,
			handler.GetProgram(
				&db.DBContext{
					Context: c,
					TLADB:   d,
				}),
		) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})

	// dbConsistencyWarning(t)

	// add this test when createProgram gets refactored
	/*
	   t.Run("TypicalRequest", func(t *testing.T) {
	           // get some random doc

	           iter := d.Collection(d.programsPath).Documents(context.Background())
	           defer iter.Stop()
	           randomDoc, err := iter.Next()
	           assert.NoError(t, err)
	           t.Logf("using doc ID (%s)", randomDoc.Ref.ID)

	           req := httptest.NewRequest(http.MethodGet, "/?pid="+randomDoc.Ref.ID, nil)
	           rec := httptest.NewRecorder()
	           assert.NotNil(t, req, rec)
	           c := echo.New().NewContext(req, rec)

	           if assert.NoError(t, handler.GetProgram(c)) {
	                   assert.Equal(t, http.StatusOK, rec.Code)
	                   assert.NotEmpty(t, rec.Result().Body)
	                   p := db.Program()
	                   bytes, err := ioutil.ReadAll(rec.Result().Body)
	                   assert.NoError(t, err)
	                   assert.NoError(t, json.Unmarshal(bytes, &p))
	                   assert.NotZero(t, p)
	                   assert.Equal(t, randomDoc.Ref.ID, p.UID) // check that the UID match
	           }
	   })
	*/
}

func TestCreateProgram(t *testing.T) {

	t.Run("BaseCase", func(t *testing.T) {
		//createUser test
		d := db.OpenMock()
		req := httptest.NewRequest(http.MethodPut, "/", nil)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		handler.CreateUser(&db.DBContext{
			Context: c,
			TLADB:   d,
		})

			// try to marshall result into an user struct
		u := db.User{}
		json.Unmarshal(rec.Body.Bytes(), &u)

		sampleDoc := struct {
			UID  string  `json:"uid"`
			Prog db.Program `json:"program"`
		}{
			UID: u.UID,
			Prog: db.Program{
				Language:  "python",
				Name:      "some random name",
				Thumbnail: 0,
			},
		}
		b, err := json.Marshal(&sampleDoc)
		require.NoError(t, err)

		req, rec = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(b))), httptest.NewRecorder()
		c = echo.New().NewContext(req, rec)
		if assert.NoError(t, handler.CreateProgram(&db.DBContext{
			Context: c,
			TLADB: d,
		}),
		){
			assert.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			assert.NotEmpty(t, rec.Result().Body)
		}

		programToDelete := d.Collection(programsPath).Doc(request.PID)
		b, err := json.Marshal(&request)
		require.NoError(t, err)

		// try to delete it
		req, rec := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(string(b))), httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)
		
		if assert.NoError(t, handler.DeleteProgram(&db.DBContext{
			Context: c,
			TLADB:   d,
		}) {
			assert.Equal(t, http.StatusOK, rec.Code)

			// check that the program actually was deleted
			_, err := programToDelete.Get(context.Background())
			assert.Equal(t, codes.NotFound, status.Code(err))
		}
	})
}

func TestDeleteProgram(t *testing.T) {
	d := db.OpenMock()
	
	t.Log("this test will assume that the first userdoc pulled from staging has at least one program")

	t.Run("TypicalRequest", func(t *testing.T) {
		randomUser := User{}

		err = d.RunTransaction(context.Background(), func(ctx context.Context, tx *firestore.Transaction) error {
			// get some random user with more than one program
			userDoc, err := d.Collection(usersPath).DocumentRefs(context.Background()).Next()
			if err != nil {
				return err
			}
			userSnap, err := tx.Get(userDoc)
			if err != nil {
				return err
			}
			return userSnap.DataTo(&randomUser)
		})
		require.NoError(t, err)

		// request struct
		request := struct {
			UID string `json:"uid"`
			PID string `json:"pid"`
		}{
			UID: randomUser.UID,
			PID: randomUser.Programs[0],
		}
		t.Logf("trying to delete program with pid (%s)", request.PID)
		programToDelete := d.Collection(programsPath).Doc(request.PID)
		b, err := json.Marshal(&request)
		require.NoError(t, err)

		// try to delete it
		req, rec := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(string(b))), httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)
		if assert.NoError(t, d.DeleteProgram(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			// check that the program actually was deleted
			_, err := programToDelete.Get(context.Background())
			assert.Equal(t, codes.NotFound, status.Code(err))
		}
	})
}