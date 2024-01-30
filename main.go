package main

import (
	"encoding/json"
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

	db.Create(&zeroFolder)
	db.Create(&testJsonObj)
	db.Create(&testFolder)
	db.Create(&testFolder3)
	db.Create(&testFolder2)
	db.Create(&testJsonObj2)


	// result := db.Raw(`
	// 	WITH RECURSIVE rectree AS (
	// 		SELECT *
	// 		FROM folders
	// 		WHERE id = 1
	// 		UNION ALL
	// 		SELECT f.*
	// 		FROM folders f
	// 		JOIN rectree
	// 		ON f.parent_id = rectree.id
	// 	) SELECT * FROM rectree;
	// `).Scan(&folders)
	var result []map[string]interface{}

	tx := db.Raw(`
	WITH RECURSIVE T1(id,name,parent_id) AS (
		SELECT * FROM folders T0 WHERE
		T0.parent_id IS 1
		UNION ALL
		SELECT T2. id, T2.name, T2.parent_id FROM folders T2, T1
		WHERE T2.parent_id = T1.id
		)
		SELECT * FROM T1;
`).Scan(&result)
	if tx.Error != nil {
		fmt.Println(tx.Error)
		return
	}
	bytes, _ := json.Marshal(result)
	fmt.Println(string(bytes))

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

	r.POST("/jsonobj", func(c *gin.Context) {
		createJSONObj(c, db)
	})
	r.GET("/jsonobj/:id", func(c *gin.Context) {
		getJSONObj(c, db)
	})
	r.PUT("/jsonobj/:id", func(c *gin.Context) {
		updateJSONObj(c, db)
	})
	r.DELETE("/jsonobj/:id", func(c *gin.Context) {
		deleteJSONObj(c, db)
	})
	r.GET("/structure", func(c *gin.Context) {
		getFolders(c, db)
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

func createJSONObj(c *gin.Context, db *gorm.DB) {
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

func getJSONObj(c *gin.Context, db *gorm.DB) {
	var jsonObj Object
	id := c.Param("id")

	result := db.First(&jsonObj, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "JSON object not found"})
		return
	}

	c.JSON(http.StatusOK, jsonObj)
}

func updateJSONObj(c *gin.Context, db *gorm.DB) {
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

func deleteJSONObj(c *gin.Context, db *gorm.DB) {
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

func getFolders(c *gin.Context, db *gorm.DB) {
	var folders []Folder

	result := db.Joins("Children").Joins("Objects").Find(&folders)
	if result.Error != nil {
		panic(result.Error)
	}

	fmt.Println(folders)

	c.JSON(http.StatusOK, []Folder{})
}
