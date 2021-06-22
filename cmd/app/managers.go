package app

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/shodikhuja83/crud/cmd/app/middleware"
	"github.com/shodikhuja83/crud/pkg/managers"
	"github.com/gorilla/mux"

)

const ADMIN = "ADMIN"

func (s *Server) handleManagerRegistration(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())

	if err != nil {
		errWriter(w, http.StatusBadRequest, err)
		return
	}

	if id == 0 {
		errWriter(w, http.StatusForbidden, err)
		return
	}

	var registrationItem struct {
		ID    int64    `json:"id"`
		Name  string   `json:"name"`
		Phone string   `json:"phone"`
		Roles []string `json:"roles"`
	}

	err = json.NewDecoder(r.Body).Decode(&registrationItem)
	if err != nil {
		errWriter(w, http.StatusInternalServerError, err)
		return
	}

	Admin := s.managerSvc.IsAdmin(r.Context(),id)
	if Admin != true {
		errWriter(w, http.StatusInternalServerError, err)
		return
	}

	item := &managers.Manager{
		ID:    registrationItem.ID,
		Name:  registrationItem.Name,
		Phone: registrationItem.Phone,
	}

	for _, role := range registrationItem.Roles {
		if role == ADMIN {
			item.IsAdmin = true
			break
		}
	}

	token, err := s.managerSvc.Create(r.Context(), item)
	if err != nil {
		errWriter(w, http.StatusInternalServerError, err)
		return
	}
	resJson(w, map[string]interface{}{"token": token})
}

func (s *Server) handleManagerGetToken(w http.ResponseWriter, r *http.Request) {
	var manager *managers.Manager
	err := json.NewDecoder(r.Body).Decode(&manager)
	if err != nil {
		errWriter(w, http.StatusInternalServerError, err)
		return
	}

	token, err := s.managerSvc.Token(r.Context(), manager.Phone, manager.Password)
	if err != nil {
		errWriter(w, http.StatusBadRequest, err)
		return
	}


	resJson(w, map[string]interface{}{"token": token})
}

func (s *Server) handleManagerChangeProducts(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())

	if err != nil {
		errWriter(w, http.StatusForbidden, err)
		return
	}

	if id == 0 {
		errWriter(w, http.StatusForbidden, err)
		return
	}

	product := &managers.Product{}
	err = json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		errWriter(w, http.StatusInternalServerError, err)
		return
	}

	product, err = s.managerSvc.SaveProduct(r.Context(),product)
	if err != nil {
		errWriter(w, http.StatusForbidden, err)
		return
	}

	resJson(w, product)
}

func (s *Server) handleManagerMakeSales(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())

	if err != nil {
		errWriter(w, http.StatusBadRequest, err)
		return
	}

	if id == 0 {
		errWriter(w, http.StatusForbidden, err)
		return
	}

	sale := &managers.Sale{}
	sale.ManagerID = id
	err = json.NewDecoder(r.Body).Decode(&sale)
	if err != nil {
		errWriter(w, http.StatusInternalServerError, err)
		return
	}

	sale, err = s.managerSvc.MakeSale(r.Context(), sale)
	if err != nil {
		errWriter(w, http.StatusBadRequest, err)
		return
	}

	resJson(w, sale)
}

func (s *Server) handleManagerGetSales(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())

	if err != nil {
		errWriter(w, http.StatusBadRequest, err)
		return
	}

	if id == 0 {
		errWriter(w, http.StatusForbidden, err)
		return
	}

	total, err := s.managerSvc.GetSales(r.Context(), id)
	if err != nil {
		errWriter(w, http.StatusBadRequest, err)
		return
	}
	resJson(w, map[string]interface{}{"manager_id": id, "total": total})
}

func (s *Server) handleManagerGetProducts(w http.ResponseWriter, r *http.Request) {
	items, err := s.managerSvc.Products(r.Context())
	if err != nil {
		errWriter(w, http.StatusBadRequest, err)
		return
	}
	

	resJson(w, items)
}

func (s *Server) handleManagerRemoveProductByID(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())

	if err != nil {
		errWriter(w, http.StatusBadRequest, err)
		return
	}

	if id == 0 {
		errWriter(w, http.StatusForbidden, err)
		return
	}

	idParam, ok := mux.Vars(r)["id"]
	if !ok {
		errWriter(w, http.StatusBadRequest, err)
		return
	}

	productID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		errWriter(w, http.StatusBadRequest, err)
		return
	}

	err = s.managerSvc.RemoveProductByID(r.Context(), productID)
	if err != nil {
		errWriter(w, http.StatusBadRequest, err)
		return
	}
}

func (s *Server) handleManagerRemoveCustomerByID(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())

	if err != nil {
		errWriter(w, http.StatusBadRequest, err)
		return
	}

	if id == 0 {
		errWriter(w, http.StatusForbidden, err)
		return
	}

	idParam, ok := mux.Vars(r)["id"]
	if !ok {
		errWriter(w, http.StatusBadRequest, err)
		return
	}

	customerID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		errWriter(w, http.StatusBadRequest, err)
		return
	}

	err = s.managerSvc.RemoveCustomerByID(r.Context(), customerID)
	if err != nil {
		errWriter(w, http.StatusBadRequest, err)
		return
	}

}

func (s *Server) handleManagerGetCustomers(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())

	if err != nil {
		errWriter(w, http.StatusBadRequest, err)
		return
	}

	if id == 0 {
		errWriter(w, http.StatusForbidden, err)
		return
	}

	items, err := s.managerSvc.Customers(r.Context())
	if err != nil {
		errWriter(w, http.StatusBadRequest, err)
		return
	}

	resJson(w, items)
}

func (s *Server) handleManagerChangeCustomer(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())

	if err != nil {
		errWriter(w, http.StatusBadRequest, err)
		return
	}

	if id == 0 {
		errWriter(w, http.StatusForbidden, err)
		return
	}
	customer := &managers.Customer{}
	err = json.NewDecoder(r.Body).Decode(&customer)
	if err != nil {
		errWriter(w, http.StatusInternalServerError, err)
		return
	}

	customer, err = s.managerSvc.ChangeCustomer(r.Context(), customer)
	if err != nil {
		errWriter(w, http.StatusBadRequest, err)
		return
	}

	resJson(w, customer)

}