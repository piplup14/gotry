package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/minio/minio-go"
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

var ids = []int{}

var images = []img{}

//GET ALL-----------------------------------------------------
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
		ids = append(ids, newItem.ID)
		images = append(images, newItem)

	}

	CheckError(err)

	c.IndentedJSON(http.StatusOK, images)
}

//ADD-----------------------------------------------------------
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
	var n = rand.Int()
	for i := 0; i < len(ids); i++ {
		if n == ids[i] {
			n = rand.Int()
		}
	}

	images = append(images, newAlbum)

	var id = rand.Int()

	result, err := db.Exec("insert into images (path, date, name, id) values ($1, $2, $3, $4)", newAlbum.Path, newAlbum.Date, newAlbum.Name, newAlbum.ID)
	CheckError(err)

	fmt.Println(result.RowsAffected())

	//minio part --------------------------------------------------------------
	endpoint := "play.minio.io:9000"
	accessKeyID := "Q3AM3UQ867SPQQA43P2F"
	secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	useSSL := true

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalln(err)
	}

	// Make a new bucket called mymusic.
	bucketName := "imagesapi"
	location := "us-east-1"

	err = minioClient.MakeBucket(bucketName, location)
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, err := minioClient.BucketExists(bucketName)
		if err == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}

	// Upload the file
	filePath := newAlbum.Path
	contentType := "image/jpeg"

	// Upload the file with FPutObject
	s, err := minioClient.FPutObject(bucketName, newAlbum.Name, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Successfully uploaded %s of size %d\n", newAlbum.Name, s)
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

//SERVER-------------------------------------------------------------
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
