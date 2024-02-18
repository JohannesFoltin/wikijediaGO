package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Folder struct {
	ID       uint `gorm:"primaryKey"`
	Name     string
	ParentID uint
	Children []Folder `gorm:"foreignKey:ParentID"`
	Objects  []Object `gorm:"foreignKey:FolderID"`
}

func createFolder(w http.ResponseWriter, r *http.Request) {
	fmt.Println("create Folder")

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

	w.WriteHeader(http.StatusOK)
}

func getFolder(w http.ResponseWriter, r *http.Request) {
	fmt.Println("getFolder")

	var folder Folder
	id := r.PathValue("id")
	fmt.Println(id)

	result := db.First(&folder, id)
	if result.Error != nil {
		http.Error(w, "Folder not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(folder)
}

func updateFolder(w http.ResponseWriter, r *http.Request) {
	fmt.Println("updateFolder")
	var folder Folder

	id := r.PathValue("id")

	err := json.NewDecoder(r.Body).Decode(&folder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := db.First(&folder, id)
	if result.Error != nil {
		http.Error(w, "Folder not found", http.StatusNotFound)
		return
	}

	result = db.Save(&folder)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteFolder(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete Folder")
	var folder Folder
	id := r.PathValue("id")

	result := db.First(&folder, id)
	if result.Error != nil {
		http.Error(w, "Folder not found", http.StatusNotFound)
		return
	}

	// Delete JSON objects in the folder
	db.Delete(&Folder{}, id)
	db.Where("folder_id = ?", id).Delete(&Object{})
	db.Where("parent_id = ?", id).Delete(&Folder{})

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Folder and its contents deleted")
}
