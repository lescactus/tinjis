package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// NewApp returns a new app instance.
// It takes in parameters the tcp address for the server to listen on
// and a pointer to a mux router.
func NewApp(addr string, r *mux.Router) *App {
	return &App{
		Server: &http.Server{
			Addr:              addr,
			Handler:           r,
			ReadTimeout:       30 * time.Second,
			ReadHeaderTimeout: 10 * time.Second,
			WriteTimeout:      30 * time.Second,
		},
	}
}

// App represents the 'payment' application.
// It comes with a custom HTTP server and can be extended
type App struct {
	// HTTP server with custom handler and read and write timeouts
	Server *http.Server
}

// Run will call ListenAndServe and wrap it in a log.Fatal() to exit
// to any error encountered
func (a *App) Run() {
	log.Println("Starting http server on port " + a.Server.Addr)
	// Bind http server configured address and start to serve requests
	log.Fatal(a.Server.ListenAndServe())
}

// Invoice represents a customer invoice
// It will be used to charge a given customer a given amount.
type Invoice struct {
	CustomerID uint64 `json:"customer_id"`

	// Currency is an empty struct as
	// Antaeus doesn't serialize the currency properly
	// and send the following example json payload:
	//
	// {"currency":{},"customer_id":1,"value":301.99}
	//
	// The 'currency' key being an empty object, hence the
	// empty Currency struct
	Currency struct{} `json:"currency,omitempty"`

	// Values represents the amount in Currency to charge to the customer.
	Value float64 `json:"value,omitempty"`

	// Result represents the success or failure of the invoice charge.
	// It can be successful or failed.
	Result bool `json:"result"`
}

// InvoiceHandler is the handler receiving the invoice charge request.
// It ensure the POST request body is valid and will try to charge the invoice.
// As per the requirements, some invoices must randomly fails
func (a *App) InvoiceHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var d Invoice
	err = json.Unmarshal(body, &d)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	d.Result = success()
	resp, err := json.Marshal(d)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, fmt.Sprintf("%v", string(resp)))
}

// healthCheck provide a simple health endpoint, typically useful
// to be used by a Kubernetes readinessProbe/livenessProbe.
//
// ref: https://inadarei.github.io/rfc-healthcheck/#
//
// The above docment provides a good starting point on what to include in a API healthcheck.
// In our case, 'payment' is a very minimal service without external dependencies
// The health check is then a very simple ping-pong.
func (a *App) healthCheck(w http.ResponseWriter, r *http.Request) {
	// Set a 200 OK response with a Content-Type http header
	w.Header().Set("Content-Type", "application/health+json")
	w.WriteHeader(http.StatusOK)

	io.WriteString(w, `{"status":"pass"}`)
}

func main() {
	r := mux.NewRouter()
	app := NewApp(":8080", r) // using port tcp/8080

	// Routes registration
	r.HandleFunc("/rest/v1/charge", app.InvoiceHandler).Methods("POST")
	r.HandleFunc("/rest/ready", app.healthCheck).Methods("GET")
	r.HandleFunc("/rest/alive", app.healthCheck).Methods("GET")

	// Using logging middleware
	r.Use(func(next http.Handler) http.Handler { return handlers.LoggingHandler(os.Stdout, next) })

	app.Run()
}

// success is a function that will randomly return true or false.
// It uses the math/rand package to generate pseudo-random numbers.
// It returns a bool.
func success() bool {
	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(2)

	if i == 0 {
		return true
	} else {
		return false
	}
}
