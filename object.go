package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Object struct {
	ID       uint `gorm:"primaryKey"`
	Name     string
	Type     string
	Data     string
	FolderID uint `gorm:"not null"`
}

func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	buf := make([]byte, 512)

	_, err = file.Read(buf)
	if err != nil {
		return
	}
	_, _ = file.Seek(0, 0)

	contentType := http.DetectContentType(buf)

	fmt.Println("contentType", contentType)

	if contentType != "image/png" && contentType != "image/jpeg" && contentType != "application/pdf" {
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

	tmpObject := Object{Name: fileHeader.Filename, Type: contentType, Data: filePath, FolderID: 1}

	result := db.Create(&tmpObject)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "File uploaded successfully")
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

	w.WriteHeader(http.StatusOK)
}

func getObject(w http.ResponseWriter, r *http.Request) {
	var jsonObj Object
	id := r.PathValue("id")

	result := db.First(&jsonObj, id)
	if result.Error != nil {
		http.Error(w, "JSON object not found", http.StatusNotFound)
		return
	}

	if jsonObj.Type != "MD" {
		jsonObj.Data = ""
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(jsonObj)
	if err != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}

}

func getFileFromObject(w http.ResponseWriter, r *http.Request) {
	var jsonObj Object
	id := r.PathValue("id")

	result := db.First(&jsonObj, id)
	if result.Error != nil {
		http.Error(w, "JSON object not found", http.StatusNotFound)
		return
	}

	if jsonObj.Type == "MD" {
		http.Error(w, "You can only get a File from an File lol. U looser", http.StatusBadRequest)
		return
	}

	filePath := filepath.FromSlash(jsonObj.Data)
	http.ServeFile(w, r, filePath)

}

func updateObject(w http.ResponseWriter, r *http.Request) {
	var jsonObj Object
	id := r.PathValue("id")

	err := json.NewDecoder(r.Body).Decode(&jsonObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if jsonObj.Type != "MD" {
		http.Error(w, "Wrong Object Type. Only MD-Files can be updated", http.StatusBadRequest)
		return
	}

	result := db.First(&jsonObj, id)
	if result.Error != nil {
		http.Error(w, "JSON object not found", http.StatusNotFound)
		return
	}

	result = db.Save(&jsonObj)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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

	w.WriteHeader(http.StatusOK)
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

	if jsonObj.Type != "MD" {
		err := os.Remove(jsonObj.Data)
		if err != nil {
			fmt.Println("Cant find Objekt?")
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Println(w, "object deleted")
}
