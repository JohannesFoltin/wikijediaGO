package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	var err error

	db, err = gorm.Open(sqlite.Open("./Data/gorm.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	tables, err := db.Migrator().GetTables()
	if err != nil {
		panic(err)
	}
	for _, table := range tables {
		fmt.Println(table)
		db.Migrator().DropTable(table)
	}

	db.AutoMigrate(&Folder{}, &Object{})

	var testJsonObj = Object{Name: "JSONTEST", Type: "MD", Data: "#asd #asd #asd", FolderID: 2}

	var zeroFolder = Folder{
		Name:     "root",
		ParentID: 0,
		Children: []Folder{},
		Objects:  []Object{},
	}
	var testFolder = Folder{Name: "TestFolder", ParentID: 1, Children: []Folder{}, Objects: []Object{}}
	var testFolder2 = Folder{Name: "TestFolder2", ParentID: 2, Children: []Folder{}, Objects: []Object{}}
	var testFolder3 = Folder{Name: "TestFolder3", ParentID: 2, Children: []Folder{}, Objects: []Object{}}
	var testJsonObj2 = Object{Name: "Test2JSOPN", Type: "MD", Data: "#asd #asdasdasd #asd", FolderID: 3}
	var testJsonObj3 = Object{Name: "Test3JSOPN", Type: "MD", Data: "#asd #asdasdasd #asd", FolderID: 3}

	db.Create(&zeroFolder)
	db.Create(&testJsonObj)
	db.Create(&testFolder)
	db.Create(&testFolder3)
	db.Create(&testFolder2)
	db.Create(&testJsonObj2)
	db.Create(&testJsonObj3)

	// Create a new folder
	http.Handle("POST /folder", enableCORS(http.HandlerFunc(createFolder)))
	http.Handle("OPTIONS /folder", http.HandlerFunc(handleCorsRequest))

	// Get a folder by ID
	http.Handle("GET /folder/{id}", enableCORS(http.HandlerFunc(getFolder)))

	// Update a folder by ID
	http.Handle("PUT /folder/{id}", enableCORS(http.HandlerFunc(updateFolder)))
	http.Handle("OPTIONS /folder/{id}", http.HandlerFunc(handleCorsRequest))

	// Delete a folder by ID
	http.Handle("DELETE /folder/{id}", enableCORS(http.HandlerFunc(deleteFolder)))

	// Create a new JSON object
	http.Handle("POST /object", enableCORS(http.HandlerFunc(createObject)))
	http.Handle("OPTIONS /object", http.HandlerFunc(handleCorsRequest))

	// Get a JSON object by ID
	http.Handle("GET /object/{id}", enableCORS(http.HandlerFunc(getObject)))
	http.Handle("GET /object/data/{id}", enableCORS(http.HandlerFunc(getFileFromObject)))

	// Update a JSON object by ID
	http.Handle("PUT /object/{id}", enableCORS(http.HandlerFunc(updateObject)))
	http.Handle("OPTIONS /object/{id}", http.HandlerFunc(handleCorsRequest))

	// Update a JSON object's name by ID
	http.Handle("PUT /object/{id}/name", enableCORS(http.HandlerFunc(updateObjectName)))
	http.Handle("OPTIONS /object/{id}/name", http.HandlerFunc(handleCorsRequest))

	// Delete a JSON object by ID
	http.Handle("DELETE /object/{id}", enableCORS(http.HandlerFunc(deleteObject)))

	// Get the folder structure
	http.Handle("GET /structure", enableCORS(http.HandlerFunc(getStructure)))

	// File upload
	http.Handle("POST /upload", enableCORS(http.HandlerFunc(handleFileUpload)))
	http.Handle("OPTIONS /upload", http.HandlerFunc(handleCorsRequest))

	// Start the server
	fmt.Println("Server started")
	errh := http.ListenAndServe(":8080", nil)
	if errh != nil {
		fmt.Println("http Server error")
		return
	}

}

func getStructure(w http.ResponseWriter, r *http.Request) {
	var folders []Folder

	// Retrieve all folders from the database
	result := db.Preload("Objects").Find(&folders)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Filter out the "Data" column from all objects
	for i := range folders {
		for j := range folders[i].Objects {
			folders[i].Objects[j].Data = ""
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(folders)
}
