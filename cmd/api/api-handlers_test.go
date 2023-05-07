package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	"web-app/pkg/data"

	"github.com/go-chi/chi/v5"
)

func Test_app_authenticate(t *testing.T) {
	var tests = []struct {
		name               string
		requestBody        string
		expectedStatusCode int
	}{
		{"valid user", `{"email":"admin@example.com","password":"secret"}`, http.StatusOK},
		{"not JSON", "I am not json", http.StatusUnauthorized},
		{"empty json", "{}", http.StatusUnauthorized},
		{"empty email", `{"email":"","password":"secret"}`, http.StatusUnauthorized},
		{"empty password", `{"email":"","password":""}`, http.StatusUnauthorized},
		{"invalid user", `{"email":"wrong@email.com","password":"wrongpassword"}`, http.StatusUnauthorized},
		{"wrong password", `{"email":"admin@example.com","password":"wrongpassword"}`, http.StatusUnauthorized},
	}

	for _, e := range tests {
		reader := strings.NewReader(e.requestBody)
		req, _ := http.NewRequest("POST", "/auth", reader)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.authenticate)
		handler.ServeHTTP(rr, req)
		if e.expectedStatusCode != rr.Code {
			t.Errorf("%s: returned wrong status code. expected %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
	}
}

func Test_app_refresh(t *testing.T) {
	var tests = []struct {
		name               string
		token              string
		expectedStatusCode int
		resetRefreshTime   bool
	}{
		{"valid", "", http.StatusOK, true},
		{"valid, but not expired", "", http.StatusTooEarly, false},
		{"expired token", expiredToken, http.StatusBadRequest, false},
	}

	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	oldRefreshTime := refreshTokenExpiry

	for _, e := range tests {
		var tkn string
		if e.token == "" {
			if e.resetRefreshTime {
				refreshTokenExpiry = time.Second * 1
			}
			tokens, _ := app.generateTokenPair(&testUser)
			tkn = tokens.RefreshToken
		} else {
			tkn = e.token
		}

		postedData := url.Values{
			"refresh_token": {tkn},
		}

		req, _ := http.NewRequest("POST", "/refresh-token", strings.NewReader(postedData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(app.refresh)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: expected status of %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
		refreshTokenExpiry = oldRefreshTime
	}
}

func Test_app_userHandlers(t *testing.T) {
	var tests = []struct {
		name           string
		method         string
		json           string
		paramID        string
		handler        http.HandlerFunc
		expectedStatus int
	}{
		{"all users", "GET", "", "", app.allUsers, http.StatusOK},
		{"delete user", "DELETE", "", "1", app.deleteUser, http.StatusNoContent},
		{"delete user bad url param", "DELETE", "", "Y", app.deleteUser, http.StatusBadRequest},
		{"get valid user", "GET", "", "1", app.getUser, http.StatusOK},
		{"get valid user bad url param", "GET", "", "Y", app.getUser, http.StatusBadRequest},
		{"get invalid user", "GET", "", "100", app.getUser, http.StatusBadRequest},
		{
			"update valid user",
			"PUT",
			`{"first_name":"Administrator","last_name":"User","email":"admin@example.com"}`,
			"1",
			app.updateUser,
			http.StatusNoContent,
		},
		{
			"update invalid user",
			"PUT",
			`{"first_name":"Administrator","last_name":"User","email":"admin@example.com"}`,
			"100",
			app.updateUser,
			http.StatusBadRequest,
		},
		{
			"update valid user - invalid json",
			"PUT",
			`{first_name:"Administrator","last_name":"User","email":"admin@example.com"}`,
			"1",
			app.updateUser,
			http.StatusBadRequest,
		},
		{
			"update valid user - bad url param",
			"PUT",
			`{first_name:"Administrator","last_name":"User","email":"admin@example.com"}`,
			"Y",
			app.updateUser,
			http.StatusBadRequest,
		},
		{
			"insert valid user",
			"POST",
			`{"first_name":"Jack","last_name":"Smith","email":"jack@example.com"}`,
			"",
			app.insertUser,
			http.StatusNoContent,
		},
		{
			"insert invalid user",
			"POST",
			`{"first_name":"Jack","last_name":"Smith","email":"jack@example.com","foo":"bar"}`,
			"",
			app.insertUser,
			http.StatusBadRequest,
		},
		{
			"insert valid user - invalid json",
			"POST",
			`{first_name:"Jack","last_name":"Smith","email":"jack@example.com"}`,
			"",
			app.insertUser,
			http.StatusBadRequest,
		},
	}

	for _, e := range tests {
		var req *http.Request
		if e.json == "" {
			req, _ = http.NewRequest(e.method, "/", nil)
		} else {
			req, _ = http.NewRequest(e.method, "/", strings.NewReader(e.json))
		}

		if e.paramID != "" {
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("userID", e.paramID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(e.handler)

		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatus {
			t.Errorf("%s: wrong status returned. expected %d, but got %d", e.name, e.expectedStatus, rr.Code)
		}
	}
}
