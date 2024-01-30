package main

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

func main() {
	db, err := gorm.Open(sqlite.Open("./Data/gorm.db"), &gorm.Config{})
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
	var testFolder3 = Folder{Name: "TestFolder3", ParentID: 3, Children: []Folder{}, Objects: []Object{}}
	var testJsonObj2 = Object{Name: "Test2JSOPN", Type: "MD", Data: "#asd #asdasdasd #asd", FolderID: 3}
	var testJsonObj3 = Object{Name: "Test3JSOPN", Type: "MD", Data: "#asd #asdasdasd #asd", FolderID: 3}

	db.Create(&zeroFolder)
	db.Create(&testJsonObj)
	db.Create(&testFolder)
	db.Create(&testFolder3)
	db.Create(&testFolder2)
	db.Create(&testJsonObj2)
	db.Create(&testJsonObj3)



	r := gin.Default()
	r.Use(cors.Default())

	r.POST("/folder", func(c *gin.Context) {
		createFolder(c, db)
	})
	r.GET("/folder/:id", func(c *gin.Context) {
		getFolder(c, db)
	})
	r.PUT("/folder/:id", func(c *gin.Context) {
		updateFolder(c, db)
	})
	r.DELETE("/folder/:id", func(c *gin.Context) {
		deleteFolder(c, db)
	})

	r.POST("/object", func(c *gin.Context) {
		createObject(c, db)
	})
	r.GET("/object/:id", func(c *gin.Context) {
		getObject(c, db)
	})
	r.PUT("/object/:id", func(c *gin.Context) {
		updateObject(c, db)
	})
	r.DELETE("/object/:id", func(c *gin.Context) {
		deleteObject(c, db)
	})
	r.GET("/structure", func(c *gin.Context) {
		getStructure(c, db)
	})

	r.Run(":8080")

}


func createFolder(c *gin.Context, db *gorm.DB) {
	var folder Folder
	if err := c.ShouldBindJSON(&folder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := db.Create(&folder)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, folder)
}

func getFolder(c *gin.Context, db *gorm.DB) {
	var folder Folder
	id := c.Param("id")

	result := db.First(&folder, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	c.JSON(http.StatusOK, folder)
}

func updateFolder(c *gin.Context, db *gorm.DB) {
	var folder Folder
	id := c.Param("id")

	result := db.First(&folder, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	if err := c.ShouldBindJSON(&folder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result = db.Save(&folder)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, folder)
}

func deleteFolder(c *gin.Context, db *gorm.DB) {
	var folder Folder
	id := c.Param("id")

	result := db.First(&folder, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	// Delete subfolders recursively
	deleteSubfolders(db, folder.ID)

	// Delete JSON objects in the folder
	db.Where("folder_id = ?", folder.ID).Delete(&Object{})

	// Delete the folder itself
	result = db.Delete(&folder)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder and its contents deleted"})
}

func deleteSubfolders(db *gorm.DB, parentID uint) {
	var subfolders []Folder
	db.Where("parent_id = ?", parentID).Find(&subfolders)

	for _, subfolder := range subfolders {
		deleteSubfolders(db, subfolder.ID)
		db.Where("folder_id = ?", subfolder.ID).Delete(&Object{})
		db.Delete(&subfolder)
	}
}

func createObject(c *gin.Context, db *gorm.DB) {
	var jsonObj Object
	if err := c.ShouldBindJSON(&jsonObj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := db.Create(&jsonObj)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonObj)
}

func getObject(c *gin.Context, db *gorm.DB) {
	var jsonObj Object
	id := c.Param("id")

	result := db.First(&jsonObj, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "JSON object not found"})
		return
	}

	c.JSON(http.StatusOK, jsonObj)
}

func updateObject(c *gin.Context, db *gorm.DB) {
	var jsonObj Object
	id := c.Param("id")

	result := db.First(&jsonObj, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "JSON object not found"})
		return
	}

	if err := c.ShouldBindJSON(&jsonObj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result = db.Save(&jsonObj)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonObj)
}

func deleteObject(c *gin.Context, db *gorm.DB) {
	var jsonObj Object
	id := c.Param("id")

	result := db.First(&jsonObj, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "JSON object not found"})
		return
	}

	result = db.Delete(&jsonObj)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "JSON object deleted"})
}

func getStructure(c *gin.Context, db *gorm.DB) {
	var folders []Folder

	// Retrieve all folders from the database
	result := db.Preload("Objects").Find(&folders)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Filter out the "Data" column from all objects
	for i := range folders {
		for j := range folders[i].Objects {
			folders[i].Objects[j].Data = ""
		}
	}

	c.JSON(http.StatusOK, folders)
}
