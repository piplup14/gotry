package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/minio/minio-go"
)

type img struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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

	rows, err := db.Query(`SELECT "name", "id" , "path" FROM "images"`)
	CheckError(err)
	var newItem img
	defer rows.Close()

	for rows.Next() {
		var name string
		var id int
		var path string

		err = rows.Scan(&name, &id, &path)
		CheckError(err)
		newItem = img{id, name, path}
		fmt.Println(newItem)
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

	images = append(images, newAlbum)

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
	s, err := minioClient.FPutObject(bucketName, newAlbum.Name+".jpg", filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Successfully uploaded %s of size %d\n", newAlbum.Name, s)
	result, err := db.Exec("insert into images (path, name) values ($1, $2)", newAlbum.Name+".jpg", newAlbum.Name)
	CheckError(err)

	fmt.Println(result.RowsAffected())
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
"name":"bame",
"path":"D:/prog/gotry/hello/public/some.png"
}
*/
