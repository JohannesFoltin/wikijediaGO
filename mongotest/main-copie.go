package main

import (
	"context"
	"io"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Folder struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Name     string             `bson:"name"`
	ParentID primitive.ObjectID `bson:"parent_id,omitempty"`
	Children []Folder           `bson:"children"`
	Objects  []Object           `bson:"objects"`
}

type Object struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Name     string             `bson:"name"`
	Type     string             `bson:"type"`
	Data     string             `bson:"data"`
	FolderID primitive.ObjectID `bson:"folder_id"`
}

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://192.168.178.33:32769"))
	if err != nil {
		panic(err)
	}

	db := client.Database("your_database_name")

	collection := db.Collection("folders")

	_, err = collection.DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}

	_, err = collection.InsertOne(context.TODO(), Folder{
		Name:     "root",
		ParentID: primitive.NilObjectID,
		Children: []Folder{},
		Objects:  []Object{},
	})
	if err != nil {
		panic(err)
	}

	_, err = collection.InsertMany(context.TODO(), []interface{}{
		Folder{Name: "TestFolder", ParentID: primitive.ObjectID{}, Children: []Folder{}, Objects: []Object{}},
		Folder{Name: "TestFolder2", ParentID: primitive.ObjectID{}, Children: []Folder{}, Objects: []Object{}},
		Folder{Name: "TestFolder3", ParentID: primitive.ObjectID{}, Children: []Folder{}, Objects: []Object{}},
	})
	if err != nil {
		panic(err)
	}

	collection = db.Collection("objects")

	_, err = collection.DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}

	_, err = collection.InsertMany(context.TODO(), []interface{}{
		Object{Name: "JSONTEST", Type: "MD", Data: "#asd #asd #asd", FolderID: primitive.ObjectID{}},
		Object{Name: "Test2JSOPN", Type: "MD", Data: "#asd #asdasdasd #asd", FolderID: primitive.ObjectID{}},
		Object{Name: "Test3JSOPN", Type: "MD", Data: "#asd #asdasdasd #asd", FolderID: primitive.ObjectID{}},
	})
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	r.Use(cors.Default())

	r.POST("/folder", func(c *gin.Context) {
		createFolder(c, collection)
	})
	r.GET("/folder/:id", func(c *gin.Context) {
		getFolder(c, collection)
	})
	r.PUT("/folder/:id", func(c *gin.Context) {
		updateFolder(c, collection)
	})
	r.DELETE("/folder/:id", func(c *gin.Context) {
		deleteFolder(c, collection)
	})
	r.POST("/upload", func(c *gin.Context) {
		uploadFile(c, db)
	})
	r.POST("/object", func(c *gin.Context) {
		createObject(c, collection)
	})
	r.GET("/object/:id", func(c *gin.Context) {
		getObject(c, collection)
	})
	r.PUT("/object/:id", func(c *gin.Context) {
		updateObject(c, collection)
	})
	r.DELETE("/object/:id", func(c *gin.Context) {
		deleteObject(c, collection)
	})
	r.GET("/structure", func(c *gin.Context) {
		getStructure(c, collection)
	})

	r.Run(":8080")

}
func uploadFile(c *gin.Context, db *mongo.Database) {
	// Get the file from the request
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()

	// Create a new GridFS bucket
	bucket, err := gridfs.NewBucket(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create a new file in the bucket
	uploadStream, err := bucket.OpenUploadStream(file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer uploadStream.Close()

	// Copy the file data to the upload stream
	_, err = io.Copy(uploadStream, src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}

func createFolder(c *gin.Context, collection *mongo.Collection) {
	var folder Folder
	if err := c.ShouldBindJSON(&folder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := collection.InsertOne(context.TODO(), folder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	folder.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusOK, folder)
}

func getFolder(c *gin.Context, collection *mongo.Collection) {
	var folder Folder
	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = collection.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&folder)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	c.JSON(http.StatusOK, folder)
}

func updateFolder(c *gin.Context, collection *mongo.Collection) {
	var folder Folder
	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = collection.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&folder)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	if err := c.ShouldBindJSON(&folder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = collection.ReplaceOne(context.TODO(), bson.M{"_id": objectID}, folder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, folder)
}

func deleteFolder(c *gin.Context, collection *mongo.Collection) {
	var folder Folder
	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = collection.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&folder)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	// Delete subfolders recursively
	deleteSubfolders(collection, folder.ID)

	// Delete JSON objects in the folder
	_, err = collection.DeleteMany(context.TODO(), bson.M{"folder_id": folder.ID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Delete the folder itself
	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder and its contents deleted"})
}

func deleteSubfolders(collection *mongo.Collection, parentID primitive.ObjectID) {
	var subfolders []Folder

	cursor, err := collection.Find(context.TODO(), bson.M{"parent_id": parentID})
	if err != nil {
		return
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var subfolder Folder
		err := cursor.Decode(&subfolder)
		if err != nil {
			continue
		}
		subfolders = append(subfolders, subfolder)
	}

	for _, subfolder := range subfolders {
		deleteSubfolders(collection, subfolder.ID)
		collection.DeleteMany(context.TODO(), bson.M{"folder_id": subfolder.ID})
		collection.DeleteOne(context.TODO(), bson.M{"_id": subfolder.ID})
	}
}

func createObject(c *gin.Context, collection *mongo.Collection) {
	var jsonObj Object
	if err := c.ShouldBindJSON(&jsonObj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := collection.InsertOne(context.TODO(), jsonObj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	jsonObj.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusOK, jsonObj)
}

func getObject(c *gin.Context, collection *mongo.Collection) {
	var jsonObj Object
	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = collection.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&jsonObj)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "JSON object not found"})
		return
	}

	c.JSON(http.StatusOK, jsonObj)
}

func updateObject(c *gin.Context, collection *mongo.Collection) {
	var jsonObj Object
	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = collection.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&jsonObj)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "JSON object not found"})
		return
	}

	if err := c.ShouldBindJSON(&jsonObj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = collection.ReplaceOne(context.TODO(), bson.M{"_id": objectID}, jsonObj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonObj)
}

func deleteObject(c *gin.Context, collection *mongo.Collection) {
	var jsonObj Object
	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = collection.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&jsonObj)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "JSON object not found"})
		return
	}

	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "JSON object deleted"})
}

func getStructure(c *gin.Context, collection *mongo.Collection) {
	var folders []Folder

	// Retrieve all folders from the database
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var folder Folder
		err := cursor.Decode(&folder)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		folders = append(folders, folder)
	}

	if err := cursor.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Filter out the "Data" field from all objects
	for i := range folders {
		for j := range folders[i].Objects {
			folders[i].Objects[j].Data = ""
		}
	}

	c.JSON(http.StatusOK, folders)
}
