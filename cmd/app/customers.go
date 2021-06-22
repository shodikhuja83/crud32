package app

import (
	"encoding/json"
	"log"
	"net/http"
	// "strings"

	"github.com/shodikhuja83/crud/cmd/app/middleware"
	"github.com/shodikhuja83/crud/pkg/customers"
	"golang.org/x/crypto/bcrypt"

)



func (s *Server) handleCustomerRegistration(w http.ResponseWriter, r *http.Request) {
	var item *customers.Registration

	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		errWriter(w, http.StatusBadRequest, err)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(item.Password), bcrypt.DefaultCost)

	if err != nil {
		errWriter(w, http.StatusInternalServerError, err)
		return
	}

	item.Password = string(hashed)

	saved, err := s.customersSvc.Register(r.Context(), item)
	if err != nil {
		errWriter(w, http.StatusInternalServerError, err)
		return
	}
	resJson(w, saved)

}
func (s *Server) handleCustomerGetToken(w http.ResponseWriter, r *http.Request) {
	var item *customers.Auth 


	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		errWriter(w, http.StatusBadRequest, err)
		return
	}

	token, err := s.customersSvc.Token(r.Context(), item.Login, item.Password)
	
	if err != nil {
		errWriter(w, http.StatusInternalServerError, err)
		return
	}


	resJson(w, map[string]interface{}{"status": "ok", "token": token})

}

func (s *Server) handleCustomerGetProducts(w http.ResponseWriter, r *http.Request) {
	items, err := s.customersSvc.Products(r.Context())
	if err != nil {
		log.Print(err)
		errWriter(w, http.StatusInternalServerError, err)
		return
	}
	resJson(w, items)
}


func (s *Server) handleCustomerGetPurchases(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())
	if err != nil {
		errWriter(w, http.StatusInternalServerError, err)
		return
	}

	items, err := s.customersSvc.Purchases(r.Context(), id)

	if err != nil {
		log.Print(err)
		errWriter(w, http.StatusInternalServerError, err)
		return
	}

	resJson(w, items)

}
