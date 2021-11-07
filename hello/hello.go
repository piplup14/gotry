package main

import (
	"net/http"

	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type img struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Date string `json:"date"`
	Path string `json:"path"`
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "admin"
	dbname   = "postgres"
)

var images = []img{}

func getAlbums(c *gin.Context) {
	images = []img{}
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// open database
	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)

	// close database
	defer db.Close()

	// check db
	err = db.Ping()
	CheckError(err)

	fmt.Println("Connected!")

	rows, err := db.Query(`SELECT "name", "id" , "date", "path" FROM "images"`)
	CheckError(err)
	var newItem img
	defer rows.Close()
	for rows.Next() {
		var name string
		var id int
		var date string
		var path string

		err = rows.Scan(&name, &id, &date, &path)
		CheckError(err)
		newItem = img{id, name, date, path}
		fmt.Println(newItem)
		images = append(images, newItem)
	}

	CheckError(err)

	c.IndentedJSON(http.StatusOK, images)
}

func postAlbums(c *gin.Context) {
	var newAlbum img

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}

	// Add the new album to the slice.
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)
	defer db.Close()

	images = append(images, newAlbum)

	result, err := db.Exec("insert into images (path, date, name, id) values ($1, $2, $3, $4)", newAlbum.Path, newAlbum.Date, newAlbum.Name, newAlbum.ID)
	CheckError(err)

	fmt.Println(result.LastInsertId())
	fmt.Println(result.RowsAffected())

}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.POST("/albums", postAlbums)

	router.Run("localhost:8980")
}

/*
{
"id":4,
"name":"bame",
"date":"01.01.1111",
"path":"a:/b/c"
}
*/
