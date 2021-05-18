/* main package
   This package will contain the main routines for the Polyglottal project
	 i.e. RESTFUL API
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Company struct {
	Id          string `json:"Id"`
	Name        string `json:"Name"`
	Homepage    string `json:"Homepage"`
	Description string `json:"Description"`
}

// Storage of all the objects
var Companies []Company

const PORT int = 8000

func seed() {
	Companies = []Company{
		{Id: "1", Name: "Google", Homepage: "https://careers.google.com/locations/tokyo/?hl=en", Description: "Very big company / FAANG"},
		{Id: "2", Name: "Degica", Homepage: "https://degica.com", Description: "Payment API specialized in Japan"},
		{Id: "3", Name: "Wealth Park", Homepage: "https://wealth-park.com", Description: "Digital solutions to property management company"},
	}
}

// -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// DATA-RELATED HANDLERS
func getAllCompanies() []Company {
	return Companies
}

func getOneCompany(id string) (*Company, error) {
	for _, company := range Companies {
		if company.Id == id {
			return &company, nil
		}
	}
	return &Company{}, fmt.Errorf("company of id=%v not found", id)
}

// -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// HTTP-QUERIES-RELATED HANDLERS

func handleAllCompanies(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit -> Download all companies")
	companiesList := getAllCompanies()
	json.NewEncoder(w).Encode(companiesList)
}

func handleSingleCompany(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	company, err := getOneCompany(vars["id"])
	if err != nil {
		// Returns a 404 if the corresponding company is not found
		http.NotFound(w, r)
		return
	}
	json.NewEncoder(w).Encode(&company)
}

// -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

func setupServer() {
	myRouter := mux.NewRouter().StrictSlash(true)
	// Defining routes for templated pages

	// api routes using a subrouter
	apiRouter := myRouter.PathPrefix("/api/v1/").Subrouter()
	apiRouter.HandleFunc("/all", handleAllCompanies)
	apiRouter.HandleFunc("/company/{id}", handleSingleCompany)

	// Starting server
	fmt.Printf("Listening on port %v\n", PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", PORT), myRouter))
}

func main() {
	seed()

	setupServer()

}
