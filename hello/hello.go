package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"time"

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
	var rightids int
	if err := c.BindJSON(&rightids); err != nil {
		return
	}

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
		if (id >= (rightids-1)*10+1) && (id <= (rightids)*10) {
			newItem = img{id, name, path}
			fmt.Println(newItem)
			images = append(images, newItem)
		}
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
	contain := strings.HasSuffix(filePath, ".jpeg")
	// Upload the file with FPutObject
	if contain {
		s, err := minioClient.FPutObject(bucketName, newAlbum.Name+".jpeg", filePath, minio.PutObjectOptions{ContentType: contentType})
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("Successfully uploaded %s of size %d\n", newAlbum.Name, s)
		result, err := db.Exec("insert into images (path, name) values ($1, $2)", newAlbum.Name+".jpg", newAlbum.Name)
		CheckError(err)
		newid, _ := result.LastInsertId()
		fmt.Println(result.RowsAffected())
		c.IndentedJSON(http.StatusOK, newid)
	} else {
		c.IndentedJSON(http.StatusBadRequest, "only jpeg")
	}
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

//get item by ID

func getOneItem(c *gin.Context) {
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
	type idForSelect struct {
		ID int `json:"id"`
	}

	var rightID idForSelect
	if err := c.BindJSON(&rightID); err != nil {
		fmt.Println("err 178")
	}

	rows, err := db.Query(`SELECT "name", "id" , "path" FROM "images" WHERE "id" = $1`, rightID.ID)
	CheckError(err)
	var newItem img
	defer rows.Close()
	//////////////
	type newurl struct {
		URL string `json:"URL"`
	}
	var saveurl newurl
	///////////////////
	for rows.Next() {
		var name string
		var id int
		var path string

		err = rows.Scan(&name, &id, &path)
		CheckError(err)
		newItem = img{id, name, path}
		fmt.Println(newItem)
		//minio part --------------------------------------------------------------
		endpoint := "play.minio.io:9000"
		accessKeyID := "Q3AM3UQ867SPQQA43P2F"
		secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
		useSSL := true

		// Initialize minio client object.
		minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
		if err != nil {
			fmt.Println("err on 209")
			log.Fatalln(err)
		}

		reqParams := make(url.Values)
		reqParams.Set("response-content-disposition", "attachment; filename=\"your-filename.txt\"")

		// Generates a presigned url which expires in a day.
		presignedURL, err := minioClient.PresignedGetObject("imagesapi", newItem.Path, time.Second*24*60*60, reqParams)
		if err != nil {
			fmt.Println(err)
			return
		}
		saveurl.URL = presignedURL.Scheme + "://" + presignedURL.Host + presignedURL.Path + "?" + presignedURL.RawQuery

		if err != nil {
			fmt.Println("err on 225")
		}
		fmt.Println("Successfully generated presigned URL", presignedURL)

		//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		c.IndentedJSON(http.StatusOK, saveurl)
	}

	CheckError(err)

}

////////////////////////     EDIT ITEM
func editItem(c *gin.Context) {
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

	var rightID img
	if err := c.BindJSON(&rightID); err != nil {
		fmt.Println("err 256")
		c.IndentedJSON(http.StatusBadRequest, "try other id")
	} else {

		rows, err := db.Query(`SELECT "name", "id" , "path" FROM "images" WHERE "id" = $1`, rightID.ID)
		CheckError(err)
		defer rows.Close()
		if !rows.Next() {
			c.IndentedJSON(http.StatusBadRequest, "id not found")
		}

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
		filePath := rightID.Path
		contentType := "image/jpeg"

		// Upload the file with FPutObject
		s, err := minioClient.FPutObject(bucketName, rightID.Name+".jpg", filePath, minio.PutObjectOptions{ContentType: contentType})
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("Successfully uploaded %s of size %d\n", rightID.Name, s)

		result, err := db.Exec("update images set path=$1, name=$2 where id = $3", rightID.Name+".jpg", rightID.Name, rightID.ID)
		CheckError(err)

		fmt.Println(result.RowsAffected())

		CheckError(err)

		c.IndentedJSON(http.StatusOK, images)
	}
}

//SERVER-------------------------------------------------------------
func main() {

	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.GET("/albums/one", getOneItem)
	router.POST("/albums/edit", editItem)

	router.POST("/albums", postAlbums)

	router.Run("localhost:8980")
}

/*

{"id":3}

{
"name":"bame",
"path":"D:/prog/gotry/hello/public/some.png"
}

{
"name":"wrongpath",
"path":"D:/prog/gotry/hello/1.jpg"
}
*/
