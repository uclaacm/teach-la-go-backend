package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
		// get some random user doc to pull UID from
		userDoc, err := d.Collection(usersPath).DocumentRefs(context.Background()).Next()
		require.NoError(t, err)
		d := db.OpenMock()

		sampleDoc := struct {
			UID  string  `json:"uid"`
			Prog Program `json:"program"`
		}{
			UID: userDoc.ID,
			Prog: Program{
				Language:  "python",
				Name:      "some random name",
				Thumbnail: 0,
			},
		}
		b, err := json.Marshal(&sampleDoc)
		require.NoError(t, err)

		req, rec := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(b))), httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)
		if assert.NoError(t, handler.CreateProgramTemp(&db.DBContext{
			context: c,
			TLADb: d,
		})) {
			assert.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			assert.NotEmpty(t, rec.Result().Body)
		}
	})
}