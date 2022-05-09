package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// To clean the test cache, run:
// 'go clean -testcache'
func TestSuccess(t *testing.T) {
	t.Run("", func(t *testing.T) {
		var numTrue, numFalse int

		for i := 0; i < 100; i++ {
			if success() {
				numTrue++
			} else {
				numFalse++
			}
		}
		assert.InDelta(t, numTrue, 50, 25)
		assert.InDelta(t, numFalse, 50, 25)
	})
}

func TestHealthCheck(t *testing.T) {
	r := mux.NewRouter()
	app := NewApp(":8080", r)

	// Create a request to pass to the handler
	req, err := http.NewRequest("GET", "/rest/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder satisfying http.ResponseWriter to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.healthCheck)
	handler.ServeHTTP(rr, req)

	expectedHTTPStatus := http.StatusOK
	expectedContentType := "application/health+json"
	expectedBody := `{"status":"pass"}`

	assert.Equal(t, expectedHTTPStatus, rr.Code)
	assert.Equal(t, expectedContentType, rr.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, rr.Body.String())
}

func TestNewApp(t *testing.T) {
	type args struct {
		addr string
		r    *mux.Router
	}
	tests := []struct {
		name string
		args args
		want *App
	}{
		{
			name: "Addr = ':8080'",
			args: args{addr: ":8080", r: mux.NewRouter()},
			want: &App{Server: &http.Server{Addr: ":8080", Handler: mux.NewRouter(), ReadTimeout: 30 * time.Second, ReadHeaderTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second}},
		},
		{
			name: "Addr = '127.0.0.1:8080'",
			args: args{addr: "127.0.0.1:8080", r: mux.NewRouter()},
			want: &App{Server: &http.Server{Addr: "127.0.0.1:8080", Handler: mux.NewRouter(), ReadTimeout: 30 * time.Second, ReadHeaderTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second}},
		},
		{
			name: "Addr = '0.0.0.0:8080'",
			args: args{addr: "0.0.0.0:8080", r: mux.NewRouter()},
			want: &App{Server: &http.Server{Addr: "0.0.0.0:8080", Handler: mux.NewRouter(), ReadTimeout: 30 * time.Second, ReadHeaderTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp(tt.args.addr, tt.args.r)
			assert.Equal(t, tt.want, app)
		})
	}
}

func TestInvoiceHandler(t *testing.T) {
	type fields struct {
		Server *http.Server
	}
	type args struct {
		body               string
		expectedBody       string
		expectedHTTPStatus int
	}

	f := fields{Server: &http.Server{}}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "Empty body",
			fields: f,
			args: args{
				body:               "",
				expectedBody:       "",
				expectedHTTPStatus: http.StatusBadRequest,
			},
		},
		{
			name:   "Invalid body",
			fields: f,
			args: args{
				body:               "this is not a valid body",
				expectedBody:       "",
				expectedHTTPStatus: http.StatusBadRequest,
			},
		},
		{
			name:   "Invalid body",
			fields: f,
			args: args{
				body:               `{"currency":{},"customer_id":"1","value":"301.99"}`,
				expectedBody:       "",
				expectedHTTPStatus: http.StatusBadRequest,
			},
		},
		{
			name:   "Valid body",
			fields: f,
			args: args{
				body:               `{"currency":{},"customer_id":1,"value":301.99}`,
				expectedBody:       `{"customer_id":1,"currency":{},"value":301.99,"result":false}`,
				expectedHTTPStatus: http.StatusOK,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/rest/v1/charge", strings.NewReader(tt.args.body))
			if err != nil {
				t.Fatal(err)
			}

			r := mux.NewRouter()
			app := NewApp(":8080", r)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(app.InvoiceHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.args.expectedHTTPStatus, rr.Code)

			if tt.args.expectedHTTPStatus == http.StatusOK {
				var req Invoice
				err = json.Unmarshal([]byte(tt.args.body), &req)
				if err != nil {
					t.Fatal(err)
				}

				var resp Invoice
				err = json.Unmarshal(rr.Body.Bytes(), &resp)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, resp.CustomerID, req.CustomerID)
				assert.Equal(t, resp.Currency, req.Currency)
				assert.Equal(t, resp.Value, req.Value)
				assert.False(t, req.Result) // default bool value
				assert.NotNil(t, resp.Result)
			} else {
				assert.Equal(t, tt.args.expectedBody, string(rr.Body.Bytes()))
			}
		})
	}
}
