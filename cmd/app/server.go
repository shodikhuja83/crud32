package app

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/shodikhuja83/crud/cmd/app/middleware"
	"github.com/shodikhuja83/crud/pkg/customers"
	"github.com/shodikhuja83/crud/pkg/managers"
	"github.com/gorilla/mux"

)

//Server ..............
type Server struct {
	mux          *mux.Router
	customersSvc *customers.Service
	managerSvc   *managers.Service
}

//NewServer: Create new Server
func NewServer(mux *mux.Router, customersSvc *customers.Service, mSvc *managers.Service) *Server {
	return &Server{
		mux:          mux,
		customersSvc: customersSvc,
		managerSvc:   mSvc,
	}
}

//function for launching handlers through mux
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

const (
	GET    = "GET"
	POST   = "POST"
	DELETE = "DELETE"
)

//Init ... server initialization
func (s *Server) Init() {

	customerAuthMd := middleware.Authenticate(s.customersSvc.IDByToken)
	customersSubrouter := s.mux.PathPrefix("/api/customers").Subrouter()

	customersSubrouter.Use(customerAuthMd)
	customersSubrouter.HandleFunc("", s.handleCustomerRegistration).Methods(POST)
	customersSubrouter.HandleFunc("/token", s.handleCustomerGetToken).Methods(POST)
	customersSubrouter.HandleFunc("/products", s.handleCustomerGetProducts).Methods(GET)
	customersSubrouter.HandleFunc("/purchases", s.handleCustomerGetPurchases).Methods(GET)

	managersAuthenticateMd := middleware.Authenticate(s.managerSvc.IDByToken)
	managersSubRouter := s.mux.PathPrefix("/api/managers").Subrouter()
	managersSubRouter.Use(managersAuthenticateMd)

	managersSubRouter.HandleFunc("", s.handleManagerRegistration).Methods(POST)
	managersSubRouter.HandleFunc("/token", s.handleManagerGetToken).Methods(POST)
	managersSubRouter.HandleFunc("/sales", s.handleManagerGetSales).Methods(GET)
	managersSubRouter.HandleFunc("/sales", s.handleManagerMakeSales).Methods(POST)
	managersSubRouter.HandleFunc("/products", s.handleManagerGetProducts).Methods(GET)
	managersSubRouter.HandleFunc("/products", s.handleManagerChangeProducts).Methods(POST)
	managersSubRouter.HandleFunc("/products/{id}", s.handleManagerRemoveProductByID).Methods(DELETE)
	managersSubRouter.HandleFunc("/customers", s.handleManagerGetCustomers).Methods(GET)
	managersSubRouter.HandleFunc("/customers", s.handleManagerChangeCustomer).Methods(POST)
	managersSubRouter.HandleFunc("/customers/{id}", s.handleManagerRemoveCustomerByID).Methods(DELETE)

}

// function for the JSON response
func resJson(w http.ResponseWriter, iData interface{}) {

	data, err := json.Marshal(iData)

	if err != nil {
		errWriter(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)

	if err != nil {

		log.Print(err)
	}
}

// function for writing an error in responseWriter
func errWriter(w http.ResponseWriter, httpSts int, err error) {
	log.Print(err)
	http.Error(w, http.StatusText(httpSts), httpSts)
}
