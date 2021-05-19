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
	"strconv"
	"text/template"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Company struct {
	gorm.Model
	Name        string `json:"Name"`
	Homepage    string `json:"Homepage"`
	Description string `json:"Description"`
}

var companyDB *gorm.DB

// -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// DATABASE SETUP RELATED

func seed(db *gorm.DB) {

	// GORM doesn't accept delete without primary key selection or WHERE constraint
	db.Where("1=1").Delete(&Company{})

	Companies := []Company{
		{Name: "Google", Homepage: "https://careers.google.com/locations/tokyo/?hl=en", Description: "Very big company / FAANG"},
		{Name: "Degica", Homepage: "https://degica.com", Description: "Payment API specialized in Japan"},
		{Name: "Wealth Park", Homepage: "https://wealth-park.com", Description: "Digital solutions to property management company"},
	}
	for _, company := range Companies {
		result := db.Select("Name", "Homepage", "Description").Create(&company)
		if result.Error != nil {
			panic("Don't manage to create records!")
		}
	}

}

func dbSetup() (db *gorm.DB) {
	db, err := gorm.Open(postgres.Open("host=localhost user=alixfachin dbname=polyglottal sslmode=disable"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("Cannot connect to the database!")
	}
	db.AutoMigrate(&Company{})
	seed(db)
	return db
}

// Premature optimization is the root of all evil -> let's get dirty first.
var homeTemplate = template.Must(template.ParseFiles("templates/base.gohtml", "templates/index.gohtml"))
var allCompaniesTemplate = template.Must(template.ParseFiles("templates/base.gohtml", "templates/allCompanies.gohtml"))

//var templates = template.Must(template.ParseFiles("templates/head.gohtml", "templates/index.gohtml", "templates/allCompanies.gohtml"))

// -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// DATA-RELATED HANDLERS
func getAllCompanies() ([]Company, error) {
	var companies []Company
	result := companyDB.Find(&companies)
	if result.Error != nil {
		return []Company{}, result.Error
	}
	return companies, nil
}

func getOneCompany(id uint) (*Company, error) {
	var company Company
	result := companyDB.First(&company, id)
	fmt.Printf("The company retrieved is %v -- result is %v \n", company, result.RowsAffected)
	if result.RowsAffected == 0 {
		return &Company{}, fmt.Errorf("company of id=%v not found", id)
	}
	return &company, nil
}

func addOneCompany(newCompany *Company) (*Company, error) {
	result := companyDB.Create(newCompany)
	if result.RowsAffected == 0 {
		return &Company{}, fmt.Errorf("failed to insert record %v into DB, %v", *newCompany, result.Error)
	}
	return newCompany, nil

}

func deleteOneCompany(companyId string) (*Company, error) {

	var companyToBeDeleted Company
	companyDB.First(&companyToBeDeleted, companyId)

	result := companyDB.Delete(&Company{}, companyId)
	if result.RowsAffected == 0 {
		return &Company{}, fmt.Errorf("failed to delete record %v into DB, %v", companyId, result.Error)
	}
	return &companyToBeDeleted, nil
}

// -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
// API-QUERIES-RELATED HANDLERS

func apiGetAllCompanies(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit -> Download all companies")
	companiesList, err := getAllCompanies()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(companiesList)
}

func apiGetSingleCompany(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit -> Download a single companies")
	vars := mux.Vars(r)
	uintID, err1 := strconv.ParseUint(vars["id"], 10, 64)
	if err1 != nil {
		http.Error(w, err1.Error(), http.StatusInternalServerError)
	}
	company, err := getOneCompany(uint(uintID))
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
	_, err = addOneCompany(&company)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(company)
}

func apiDeleteOneCompany(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DELETE Endpoint hit -> Delete a single company")
	vars := mux.Vars(r)
	company, err := deleteOneCompany(vars["id"])
	if err != nil {
		http.NotFound(w, r)
		return
	}
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
	var err error
	allCompaniesData.AllCompanies, err = getAllCompanies()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = allCompaniesTemplate.ExecuteTemplate(w, "mainPage", &allCompaniesData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// -=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
const PORT int = 8000

func setupServer() {
	myRouter := mux.NewRouter().StrictSlash(true)
	// Defining routes for templated pages
	myRouter.HandleFunc("/", handleRootPage)
	myRouter.HandleFunc("/allCompanies", handleAllCompaniesPage)

	// api routes using a subrouter
	apiRouter := myRouter.PathPrefix("/api/v1/").Subrouter()
	apiRouter.HandleFunc("/all", apiGetAllCompanies)
	apiRouter.HandleFunc("/company/{id}", apiGetSingleCompany).Methods("GET")
	apiRouter.HandleFunc("/company/{id}", apiDeleteOneCompany).Methods("DELETE")
	apiRouter.HandleFunc("/company", apiCreateNewCompany).Methods("POST")

	// Static files
	fs := http.FileServer(http.Dir("assets/"))
	myRouter.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Starting server
	fmt.Printf("Listening on port %v\n", PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", PORT), myRouter))
}

func main() {
	companyDB = dbSetup()
	setupServer()

}
