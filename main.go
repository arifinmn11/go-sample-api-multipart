package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/fairyhunter13/materi-rakamin/pkg/config"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Storage Storage `mapstructure:"STORAGE"`
}

type Storage struct {
	Endpoint  string `mapstructure:"ENDPOINT"`
	AccessKey string `mapstructure:"ACCESS_KEY"`
	SecretKey string `mapstructure:"SECRET_KEY"`
	Region    string `mapstructure:"REGION"`
	Bucket    string `mapstructure:"BUCKET"`
}

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("test")
	})

	app.Post("/", handleFileupload)

	app.Listen(":3000")
}

func handleFileupload(c *fiber.Ctx) error {

	// parse incomming image file
	file, err := c.FormFile("uploads")

	// hande error ketika file kosong
	if err != nil {
		log.Println("image upload error --> ", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})

	}

	// generate uuid
	// menganti nama dari file image
	// extract file extension dari nama file yg d upload
	// generate image from filename and extension
	uniqueId := uuid.New()
	filename := strings.Replace(uniqueId.String(), "-", "", -1)
	fileExt := strings.Split(file.Filename, ".")[1]
	image := fmt.Sprintf("%s.%s", filename, fileExt)

	//konfigurasi minio
	cfg := new(Config)
	errs := config.LoadConfig(cfg, ".env", ".env")
	if err != nil {
		log.Printf("Error in loading the config: %v.", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Error in loading the config: %v.", "error": errs})
	}

	client, err := minio.New(cfg.Storage.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV2(cfg.Storage.AccessKey, cfg.Storage.SecretKey, ""),
		Secure: true,
		Region: cfg.Storage.Region,
	})

	if err != nil {
		log.Printf("Error in initializing the new client: %v.", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Error in loading the config: %v.", "error": err})

	}

	ctx := context.Background()
	isExist, err := client.BucketExists(ctx, cfg.Storage.Bucket)
	if err != nil {
		log.Printf("Error in checking the bucket: %v.", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Error in checking the bucket:" + err.Error()})
	}

	if !isExist {
		log.Printf("Bucket %s is not exist!", cfg.Storage.Bucket)
		return c.JSON(fiber.Map{"status": 500, "message": "Bucket %s is not exist!: %v.", "error": cfg.Storage.Bucket})
	}

	isObjectExist := true
	// uploadfile, err := file.Open()
	objectInfo, err := client.StatObject(ctx, cfg.Storage.Bucket, filename, minio.GetObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code != "NoSuchKey" {
			log.Printf("Error in getting the object info: %v.", err)
			return c.JSON(fiber.Map{"status": 500, "message": "Error in getting the object info: %v.", "error": cfg.Storage.Bucket})

		}
		err = nil
		isObjectExist = false
	}

	if isObjectExist {
		fmt.Println("ObjectInfo:")
		fmt.Printf("%+v\n", objectInfo)
	}

	strReader, err := file.Open()
	_, err = client.PutObject(ctx, cfg.Storage.Bucket, image, strReader, file.Size, minio.PutObjectOptions{})
	if err != nil {
		log.Printf("Error in uploading the file #%s: %v.", image, err)
		return c.JSON(fiber.Map{"status": 500, "message": "Error in uploading the file #%s: %v.", "error": cfg.Storage.Bucket})
	}

	// create meta data and send to client
	data := map[string]interface{}{
		"fileName": image,
	}

	return c.JSON(fiber.Map{"status": 201, "message": "Image uploaded successfully", "data": data})
}

// func uploadFileToCloud() {
// 	cfg := new(Config)
// 	err := config.LoadConfig(cfg, ".env", ".env")
// 	if err != nil {
// 		log.Printf("Error in loading the config: %v.", err)
// 		return
// 	}

// 	client, err := minio.New(cfg.Storage.Endpoint, &minio.Options{
// 		Creds:  credentials.NewStaticV2(cfg.Storage.AccessKey, cfg.Storage.SecretKey, ""),
// 		Secure: true,
// 		Region: cfg.Storage.Region,
// 	})

// 	if err != nil {
// 		log.Printf("Error in initializing the new client: %v.", err)
// 		return
// 	}

// 	ctx := context.Background()
// 	isExist, err := client.BucketExists(ctx, cfg.Storage.Bucket)
// 	if err != nil {
// 		log.Printf("Error in checking the bucket: %v.", err)
// 		return
// 	}

// 	if !isExist {
// 		log.Printf("Bucket %s is not exist!", cfg.Storage.Bucket)
// 		return
// 	}

// 	isObjectExist := true
// 	objectInfo, err := client.StatObject(ctx, cfg.Storage.Bucket, testFilename, minio.GetObjectOptions{})
// 	if err != nil {
// 		errResp := minio.ToErrorResponse(err)
// 		if errResp.Code != "NoSuchKey" {
// 			log.Printf("Error in getting the object info: %v.", err)
// 			return
// 		}
// 		err = nil
// 		isObjectExist = false
// 	}

// 	if isObjectExist {
// 		fmt.Println("ObjectInfo:")
// 		fmt.Printf("%+v\n", objectInfo)
// 	}

// 	if isObjectExist {
// 		obj, err := client.GetObject(ctx, cfg.Storage.Bucket, testFilename, minio.GetObjectOptions{})
// 		if err != nil {
// 			log.Printf("Error in getting the object: %v.", err)
// 			return
// 		}

// 		sb := new(strings.Builder)
// 		_, err = io.Copy(sb, obj)
// 		if err != nil {
// 			log.Printf("Error in copying from the object reader: %v.", err)
// 			return
// 		}

// 		fmt.Printf("File \"%s\":\n", testFilename)
// 		fmt.Println(sb.String())
// 		return
// 	}

// 	strReader := strings.NewReader(testJSON)
// 	uploadInfo, err := client.PutObject(ctx, cfg.Storage.Bucket, testFilename, strReader, int64(strReader.Len()), minio.PutObjectOptions{})
// 	if err != nil {
// 		log.Printf("Error in uploading the file #%s: %v.", testFilename, err)
// 		return
// 	}

// 	log.Printf("Uploading the file #%s succeeded!", testFilename)
// 	fmt.Println("UploadInfo:")
// 	fmt.Printf("%+v\n", uploadInfo)
// }
