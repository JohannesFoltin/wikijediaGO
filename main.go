package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Folder struct {
	ID       uint `gorm:"primaryKey"`
	Name     string
	ParentID uint
	Children []Folder `gorm:"foreignKey:ParentID"`
	Objects  []Object `gorm:"foreignKey:FolderID"`
}

type Object struct {
	ID       uint `gorm:"primaryKey"`
	Name     string
	Type     string
	Data     string
	FolderID uint `gorm:"not null"`
}

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
	http.Handle("GET /posts/{id}", enableCORS(http.HandlerFunc(getFolder)))

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
	errh := http.ListenAndServe(":8080", nil)
	if errh != nil {
		fmt.Println("http Server error")
		return
	}

}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handleCorsRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(http.StatusOK)
}

func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check file type
	fileType := fileHeader.Header.Get("Content-Type")
	fmt.Println(fileType)
	if fileType != "image/png" && fileType != "image/jpeg" && fileType != "application/pdf" {
		http.Error(w, "Invalid file type", http.StatusBadRequest)
		return
	}

	// Save file to disk
	filePath := "uploads/" + fileHeader.Filename
	outFile, err := os.Create(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpObject := Object{Name: fileHeader.Filename, Type: fileType, Data: filePath, FolderID: 1}

	result := db.Create(&tmpObject)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "File uploaded successfully")
}

func createFolder(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Anfrage lol")

	var folder Folder
	err := json.NewDecoder(r.Body).Decode(&folder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := db.Create(&folder)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(folder)
}

func getFolder(w http.ResponseWriter, r *http.Request) {
	var folder Folder
	id := r.PathValue("id")

	result := db.First(&folder, id)
	if result.Error != nil {
		http.Error(w, "Folder not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(folder)
}

func updateFolder(w http.ResponseWriter, r *http.Request) {
	var folder Folder
	id := r.PathValue("id")

	result := db.First(&folder, id)
	if result.Error != nil {
		http.Error(w, "Folder not found", http.StatusNotFound)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&folder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result = db.Save(&folder)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(folder)
}

func deleteFolder(w http.ResponseWriter, r *http.Request) {
	var folder Folder
	id := r.PathValue("id")

	result := db.First(&folder, id)
	if result.Error != nil {
		http.Error(w, "Folder not found", http.StatusNotFound)
		return
	}

	// Delete subfolders recursively
	deleteSubfolders(folder.ID)

	// Delete JSON objects in the folder
	db.Where("folder_id = ?", folder.ID).Delete(&Object{})

	// Delete the folder itself
	result = db.Delete(&folder)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Folder and its contents deleted")
}

func deleteSubfolders(parentID uint) {
	var subfolders []Folder
	db.Where("parent_id = ?", parentID).Find(&subfolders)

	for _, subfolder := range subfolders {
		deleteSubfolders(subfolder.ID)
		db.Where("folder_id = ?", subfolder.ID).Delete(&Object{})
		db.Delete(&subfolder)
	}
}

func createObject(w http.ResponseWriter, r *http.Request) {
	var jsonObj Object
	err := json.NewDecoder(r.Body).Decode(&jsonObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := db.Create(&jsonObj)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonObj)
}

func getObject(w http.ResponseWriter, r *http.Request) {
	var jsonObj Object
	id := r.PathValue("id")

	result := db.First(&jsonObj, id)
	if result.Error != nil {
		http.Error(w, "JSON object not found", http.StatusNotFound)
		return
	}

	// Check the object type
	switch jsonObj.Type {
	case "MD":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jsonObj)
		return
	case "image/png":
		filePath := filepath.FromSlash(jsonObj.Data)
		http.ServeFile(w, r, filePath)
	case "jpeg":
		// Handle JPEG object
		// ...
	case "pdf":
		// Handle PDF object
		// ...
	default:
		http.Error(w, "Unsupported object type", http.StatusBadRequest)
		return
	}

	// Object handling logic
	// ...
}

func updateObject(w http.ResponseWriter, r *http.Request) {
	var jsonObj Object
	id := r.PathValue("id")

	result := db.First(&jsonObj, id)
	if result.Error != nil {
		http.Error(w, "JSON object not found", http.StatusNotFound)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&jsonObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result = db.Save(&jsonObj)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonObj)
}

func updateObjectName(w http.ResponseWriter, r *http.Request) {
	var object Object
	id := r.PathValue("id")

	result := db.First(&object, id)
	if result.Error != nil {
		http.Error(w, "Object not found", http.StatusNotFound)
		return
	}

	var updatedData struct {
		Name string `json:"name"`
	}

	err := json.NewDecoder(r.Body).Decode(&updatedData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(updatedData.Name)
	object.Name = updatedData.Name

	result = db.Save(&object)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(object)
}

func deleteObject(w http.ResponseWriter, r *http.Request) {
	var jsonObj Object
	id := r.PathValue("id")

	result := db.First(&jsonObj, id)
	if result.Error != nil {
		http.Error(w, "JSON object not found", http.StatusNotFound)
		return
	}

	result = db.Delete(&jsonObj)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "JSON object deleted")
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
