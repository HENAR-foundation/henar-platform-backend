package static

import (
	"fmt"
	"henar-backend/sentry"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var s3Client *s3.S3

func Init() error {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials("DO009TG799ZCZG7WCBHU", "ok6N6/xDW2BsLas+HG4aMI5rBZOt6Krhr0djzGSAclg", ""),
		Endpoint:         aws.String("https://ams3.digitaloceanspaces.com"),
		Region:           aws.String("ams3"),
		S3ForcePathStyle: aws.Bool(false),
	}

	newSession, err := session.NewSession(s3Config)
	if err != nil {
		sentry.SentryHandler(err)
		fmt.Println(err.Error())
	}
	s3Client = s3.New(newSession)

	return nil
}
