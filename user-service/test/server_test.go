package test

import (
	"bytes"
	// "context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	// "github.com/aws/aws-sdk-go/service/s3"
	// "github.com/opensearch-project/opensearch-go"
	"github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/mock"

	"github.com/sebaraj/crush/user-service/mocks"
	"github.com/sebaraj/crush/user-service/server"
)

type testServer struct {
	server *server.Server
	db     *sql.DB
	dbMock sqlmock.Sqlmock
	osMock *mocks.MockOpenSearchClient
	s3Mock *mocks.MockS3Client
}

func setupTestServer(t *testing.T) *testServer {
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create db mock: %v", err)
	}

	osMock := new(mocks.MockOpenSearchClient)

	s3Mock := new(mocks.MockS3Client)

	os.Setenv("OAUTH_CLIENT", "test-client-id")

	// need to pass in s3/opensearch mocks
	s := server.NewServer(db, "test-bucket", "us-east-1", nil, nil)

	return &testServer{
		server: s,
		db:     db,
		dbMock: dbMock,
		osMock: osMock,
		s3Mock: s3Mock,
	}
}

// func TestUserSearch(t *testing.T) {
// 	ts := setupTestServer(t)
// 	defer ts.db.Close()
//
// 	t.Run("successful search", func(t *testing.T) {
// 		mockResponse := &opensearch.Response{
// 			StatusCode: 200,
// 			Body: bytes.NewReader([]byte(`{
//                 "hits": {
//                     "hits": [
//                         {
//                             "_source": {
//                                 "email": "test@yale.edu",
//                                 "name": "Test User"
//                             }
//                         }
//                     ]
//                 }
//             }`)),
// 		}
//
// 		ts.osMock.On("Search", mock.Anything).Return(mockResponse, nil)
//
// 		req := httptest.NewRequest("GET", "/v1/user/search/", bytes.NewReader([]byte(`{"query": "test"}`)))
// 		req.Header.Set("Authorization", "valid-token")
// 		w := httptest.NewRecorder()
//
// 		ts.server.HandleSearch(w, req)
//
// 		assert.Equal(t, http.StatusOK, w.Code)
// 		ts.osMock.AssertExpectations(t)
// 	})
// }

func TestUserAnswers(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.db.Close()

	t.Run("update answers unauthorized", func(t *testing.T) {
		ts.dbMock.ExpectBegin()
		ts.dbMock.ExpectExec("UPDATE answers SET").
			WithArgs(5, "test@yale.edu").
			WillReturnResult(sqlmock.NewResult(1, 1))
		ts.dbMock.ExpectCommit()

		body := map[string]interface{}{
			"question1": 5,
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("PUT", "/v1/user/answers/test@yale.edu", bytes.NewReader(jsonBody))
		req.Header.Set("Authorization", "valid-token")
		w := httptest.NewRecorder()

		ts.server.HandleAnswers(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		// assert.NoError(t, ts.dbMock.ExpectationsWereMet())
	})

	// t.Run("update answers successfully", func(t *testing.T) {
	// 	ts.dbMock.ExpectBegin()
	// 	ts.dbMock.ExpectExec("UPDATE answers SET").
	// 		WithArgs(5, "test@yale.edu").
	// 		WillReturnResult(sqlmock.NewResult(1, 1))
	// 	ts.dbMock.ExpectCommit()
	//
	// 	body := map[string]interface{}{
	// 		"question1": 5,
	// 	}
	// 	jsonBody, _ := json.Marshal(body)
	// 	req := httptest.NewRequest("PUT", "/v1/user/answers/test@yale.edu", bytes.NewReader(jsonBody))
	// 	req.Header.Set("Authorization", "valid-token")
	// 	w := httptest.NewRecorder()
	//
	// 	ts.server.HandleAnswers(w, req)
	//
	// 	assert.Equal(t, http.StatusOK, w.Code)
	// 	assert.NoError(t, ts.dbMock.ExpectationsWereMet())
	// })
}

// func TestUserPicture(t *testing.T) {
// 	ts := setupTestServer(t)
// 	defer ts.db.Close()
//
// 	t.Run("get picture URL successfully", func(t *testing.T) {
// 		mockRequest := &s3.Request{}
// 		mockOutput := &s3.GetObjectOutput{}
// 		ts.s3Mock.On("GetObjectRequest", mock.Anything).Return(mockRequest, mockOutput)
//
// 		ts.dbMock.ExpectExec("UPDATE users SET picture_s3_url").
// 			WithArgs(mock.Anything, "test@yale.edu").
// 			WillReturnResult(sqlmock.NewResult(1, 1))
//
// 		req := httptest.NewRequest("GET", "/v1/user/picture/test@yale.edu", nil)
// 		req.Header.Set("Authorization", "valid-token")
// 		w := httptest.NewRecorder()
//
// 		ts.server.HandlePicture(w, req)
//
// 		assert.Equal(t, http.StatusOK, w.Code)
// 		ts.s3Mock.AssertExpectations(t)
// 		assert.NoError(t, ts.dbMock.ExpectationsWereMet())
// 	})
// }

// func TestUserInfo(t *testing.T) {
// 	ts := setupTestServer(t)
// 	defer ts.db.Close()
//
// 	t.Run("get user info successfully", func(t *testing.T) {
// 		rows := sqlmock.NewRows([]string{
// 			"email", "is_active", "name", "residential_college", "graduating_year",
// 			"gender", "partner_genders", "instagram", "snapchat", "phone_number",
// 			"picture_s3_url",
// 			"interest_1", "interest_2", "interest_3", "interest_4", "interest_5",
// 			"question1", "question2", "question3", "question4", "question5",
// 			"question6", "question7", "question8", "question9", "question10",
// 			"question11", "question12",
// 		}).AddRow(
// 			"test@yale.edu", true, "Test User", "Berkeley", 2024,
// 			1, 2, "insta", "snap", "1234567890",
// 			"https://s3.url",
// 			"Music", "Art", "Sports", nil, nil,
// 			1, 2, 3, 4, 5,
// 			1, 2, 3, 4, 5,
// 			1, 2,
// 		)
//
// 		ts.dbMock.ExpectBegin()
// 		ts.dbMock.ExpectQuery("SELECT (.+) FROM users").
// 			WithArgs("test@yale.edu").
// 			WillReturnRows(rows)
// 		ts.dbMock.ExpectCommit()
//
// 		req := httptest.NewRequest("GET", "/v1/user/info/test@yale.edu", nil)
// 		req.Header.Set("Authorization", "valid-token")
// 		w := httptest.NewRecorder()
//
// 		ts.server.HandleUser(w, req)
//
// 		assert.Equal(t, http.StatusOK, w.Code)
// 		assert.NoError(t, ts.dbMock.ExpectationsWereMet())
//
// 		var response map[string]interface{}
// 		err := json.NewDecoder(w.Body).Decode(&response)
// 		assert.NoError(t, err)
// 		assert.Equal(t, "test@yale.edu", response["email"])
// 	})
// }
