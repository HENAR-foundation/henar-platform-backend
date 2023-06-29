package static

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"henar-backend/sentry"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gofiber/fiber/v2"
)

// @Summary Upload file
// @Description Upload a static file to Henar DigitalOcean failopoika's and get the uri
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Success 200 {array} types.FileResponce
// @Failure 400 {string} string "error reading file"
// @Router /v1/files/upload [post]
func UploadFile(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		sentry.SentryHandler(err)
		c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"msg": "error reading file",
			"err": err,
		})
	}

	fmt.Println(file.Filename)

	buffer, err := file.Open()
	if err != nil {
		sentry.SentryHandler(err)
		c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"msg": "error reading file",
			"err": err,
		})
	}
	defer buffer.Close()

	fileNameSplit := strings.Split(file.Filename, ".")
	fileExt := "." + fileNameSplit[len(fileNameSplit)-1]
	fileNameMD5 := sha1.Sum([]byte(file.Filename + string(time.Now().String())))
	fileNameHashString := base64.StdEncoding.EncodeToString(fileNameMD5[:])
	fileNameFull := string(fileNameHashString) + fileExt

	object := s3.PutObjectInput{
		Bucket: aws.String("henar-static"),
		Key:    aws.String(fileNameFull),
		Body:   buffer,
		ACL:    aws.String("public-read"),
	}
	_, err = s3Client.PutObject(&object)
	if err != nil {
		sentry.SentryHandler(err)
		fmt.Println(err.Error())
	}

	c.Status(http.StatusOK).JSON(fiber.Map{
		"url": "https://henar-static.ams3.digitaloceanspaces.com/" + fileNameFull,
	})

	return nil
}
