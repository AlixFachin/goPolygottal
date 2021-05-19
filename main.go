/* main package
   This package will contain the main routines for the Polyglottal project
	 i.e. RESTFUL API
*/

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"

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

// Premature optimization is the root of all evil -> let's get dirty first.
var homeTemplate = template.Must(template.ParseFiles("templates/base.gohtml", "templates/index.gohtml"))
var allCompaniesTemplate = template.Must(template.ParseFiles("templates/base.gohtml", "templates/allCompanies.gohtml"))

//var templates = template.Must(template.ParseFiles("templates/head.gohtml", "templates/index.gohtml", "templates/allCompanies.gohtml"))

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

func addOneCompany(newCompany *Company) {
	Companies = append(Companies, *newCompany)
}

// -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// API-QUERIES-RELATED HANDLERS

func handleAllCompanies(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit -> Download all companies")
	companiesList := getAllCompanies()
	json.NewEncoder(w).Encode(companiesList)
}

func handleSingleCompany(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit -> Download a single companies")
	vars := mux.Vars(r)
	company, err := getOneCompany(vars["id"])
	if err != nil {
		// Returns a 404 if the corresponding company is not found
		http.NotFound(w, r)
		return
	}
	json.NewEncoder(w).Encode(&company)
}

func apiCreateNewCompany(w http.ResponseWriter, r *http.Request) {
	var company Company
	fmt.Println("POST request - create a new company")

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	json.Unmarshal(reqBody, &company)
	addOneCompany(&company)
	json.NewEncoder(w).Encode(company)

}

// -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// TEMPLATED PAGES

func handleRootPage(w http.ResponseWriter, r *http.Request) {
	err := homeTemplate.ExecuteTemplate(w, "mainPage", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleAllCompaniesPage(w http.ResponseWriter, r *http.Request) {
	type AllCompaniesDataType struct {
		AllCompanies []Company
	}
	var allCompaniesData AllCompaniesDataType
	allCompaniesData.AllCompanies = getAllCompanies()

	err := allCompaniesTemplate.ExecuteTemplate(w, "mainPage", &allCompaniesData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

func setupServer() {
	myRouter := mux.NewRouter().StrictSlash(true)
	// Defining routes for templated pages
	myRouter.HandleFunc("/", handleRootPage)
	myRouter.HandleFunc("/allCompanies", handleAllCompaniesPage)

	// api routes using a subrouter
	apiRouter := myRouter.PathPrefix("/api/v1/").Subrouter()
	apiRouter.HandleFunc("/all", handleAllCompanies)
	apiRouter.HandleFunc("/company/{id}", handleSingleCompany)
	apiRouter.HandleFunc("/company", apiCreateNewCompany).Methods("POST")

	// Static files
	fs := http.FileServer(http.Dir("assets/"))
	myRouter.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Starting server
	fmt.Printf("Listening on port %v\n", PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", PORT), myRouter))
}

func main() {
	seed()

	setupServer()

}
