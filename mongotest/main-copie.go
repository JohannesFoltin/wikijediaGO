package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type File struct {
	Name     string `bson:"name"`
	IsGridFS bool   `bson:"isGridFS"`
}

type Folder struct {
	Name       string   `bson:"name"`
	Files      []File   `bson:"files"`
	SubFolders []Folder `bson:"subFolders"`
}

func main() {
	// Verbindung zur MongoDB-Datenbank herstellen
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://192.168.178.33:32769"))
	if err != nil {
		panic(err)
	}

	// Datenbank und Collection auswählen
	collection := client.Database("hallejuhla").Collection("folders")

	// Beispielstruktur erstellen
	folder := Folder{
		Name: "root",
		Files: []File{
			{Name: "file1", IsGridFS: false},
			{Name: "file2", IsGridFS: true},
		},
		SubFolders: []Folder{
			{
				Name: "subFolder1",
				Files: []File{
					{Name: "file3", IsGridFS: false},
				},
			},
		},
	}

	// Beispielstruktur in die Datenbank einfügen
	_, err = collection.InsertOne(context.TODO(), folder)
	if err != nil {
		log.Fatal(err)
	}

	// Beispielstruktur als JSON ausgeben
	jsonResult, err := bson.MarshalExtJSON(folder, true, false)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(jsonResult))
}
