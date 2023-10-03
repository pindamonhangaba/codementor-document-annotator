package s3

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gabriel-vasile/mimetype"
	"github.com/pkg/errors"
)

type s3Acl string

func (acl s3Acl) ToPtr() *s3Acl {
	return &acl
}

const (
	BucketPrivateMediaFiles = "monoboi-private"
	BucketPublicMediaFiles  = "monoboi-public"
	// Owner gets FULL_CONTROL. No one else has access rights (default).
	S3ACLPrivate = s3Acl("private")
	// Owner gets FULL_CONTROL. The AllUsers group gets READ access.
	S3ACLPublicRead = s3Acl("public-read")
	// Owner gets FULL_CONTROL. The AllUsers group gets READ and WRITE access. Granting this on a bucket is generally not recommended.
	S3ACLPublicReadWrite = s3Acl("public-read-write")
)

type S3Client struct {
	client *s3.S3
}

func NewS3Client(region, accessKey, secretKey, endpoint string) *S3Client {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:    aws.String(endpoint),
	}))

	_, err := sess.Config.Credentials.Get()
	if err != nil {
		panic(err)
	}

	client := s3.New(sess)
	s3c := S3Client{client: client}
	return &s3c
}

func (s3c *S3Client) ListAllBuckets() (ret *s3.ListBucketsOutput, err error) {
	ret, err = s3c.client.ListBuckets(&s3.ListBucketsInput{})
	return ret, err
}

// Bucket names are unique to their location and must meet the following criteria:
// Only lowercase and starts with a letter or number. No spaces.
// Bucket name may contain dashes
// Must be between 3 and 63 characters long.
func (s3c *S3Client) CreateBucket(bucketName string, acl *s3Acl) (ret *s3.CreateBucketOutput, exist bool, err error) {
	exist = false

	if !nameCheck(bucketName) {
		return nil, exist, errors.New(bucketName + "does not meet the criteria.")
	}

	bucketList, err := s3c.ListAllBuckets()
	if err != nil {
		return nil, exist, err
	}
	for _, bucket := range bucketList.Buckets {
		if bucketName == *bucket.Name {
			exist = true
		}
	}

	if !exist {
		ret, err = newBucket(s3c.client, bucketName, acl)
		if err != nil {
			return nil, exist, err
		}
	}

	return ret, exist, nil
}

func nameCheck(str string) bool {
	checker := regexp.MustCompile(`^[a-z0-9-]*$`).MatchString(str)
	alphanumeric := "abcdefghijklmnopqrstuvwxyz1234567890"
	if checker && strings.Contains(alphanumeric, string(str[0])) && len(str) >= 3 && len(str) <= 63 {
		return true
	} else {
		return false
	}
}

func newBucket(S3Client *s3.S3, bucketName string, acl *s3Acl) (*s3.CreateBucketOutput, error) {
	input := s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(""),
		},
	}
	if acl != nil {
		input.ACL = aws.String(string(*acl))
	}
	ret, err := S3Client.CreateBucket(&input)
	if awsError, ok := err.(awserr.Error); ok {
		if awsError.Code() == s3.ErrCodeBucketAlreadyExists {
			log.Fatalf("Bucket %q already exists. Error: %v", bucketName, awsError.Code())
		}
		return nil, err
	}
	return ret, nil
}

func (s3c *S3Client) UploadObject(filePath, directory, bucketName string, acl *s3Acl) (string, *s3.PutObjectOutput, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", nil, err
	}
	filename := strings.Split(filePath, "/")[strings.Count(filePath, "/")]
	mtype, err := mimetype.DetectFile(filePath)
	if err != nil {
		return "", nil, err
	}

	input := s3.PutObjectInput{
		Body:        f,
		Bucket:      aws.String(bucketName),
		Key:         aws.String(path.Join(directory, filename)),
		ContentType: aws.String(mtype.String()),
	}
	if acl != nil {
		input.ACL = aws.String(string(*acl))
	}
	ret, err := s3c.client.PutObject(&input)
	if err != nil {
		return "", nil, err
	}
	req, _ := s3c.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(path.Join(directory, filename)),
	})
	req.Build()
	url := req.HTTPRequest.URL.String()
	return url, ret, nil
}

func (s3c *S3Client) ListObjects(bucketName string) (*s3.ListObjectsV2Output, error) {
	ret, err := s3c.client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})

	return ret, err
}

func (s3c *S3Client) GetSignedURL(bucketName, file string, ttl time.Duration) (string, error) {
	req, _ := s3c.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(file),
	})

	url, err := req.Presign(ttl)
	if err != nil {
		return "", errors.Wrap(err, "presign GetObjectRequest for key "+file)
	}
	return url, nil
}

func (s3c *S3Client) DeleteObject(bucketName, directory, filename string) (*s3.DeleteObjectOutput, error) {
	ret, err := s3c.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(path.Join(directory, strings.Split(filename, "/")[strings.Count(filename, "/")])),
	})

	return ret, err
}

func (s3c *S3Client) DeleteAllObjects(bucketName string) error {
	objectList, err := s3c.ListObjects(bucketName)
	if err != nil {
		return err
	}
	for _, object := range objectList.Contents {
		_, err := s3c.client.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(*object.Key),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s3c *S3Client) DeleteBucket(bucketName string) (*s3.DeleteBucketOutput, error) {
	err := s3c.DeleteAllObjects(bucketName)
	if err != nil {
		return nil, err
	}

	ret, err := s3c.client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		if awsError, ok := err.(awserr.Error); ok {
			if awsError.Code() == s3.ErrCodeNoSuchBucket {
				return nil, errors.New("No Bucket exists with the name '" + bucketName + "'")
			}
			return nil, err
		}
	}

	return ret, nil
}

func (s3c *S3Client) PutBucketCors(bucketName string, methods ...string) error {
	validMethods := aws.StringSlice(filterMethods(methods))
	if len(validMethods) == 0 {
		return errors.New(fmt.Sprintf("invalid methods, %s", strings.Join(methods, ", ")))
	}

	rule := s3.CORSRule{
		AllowedHeaders: aws.StringSlice([]string{"Authorization"}),
		AllowedOrigins: aws.StringSlice([]string{"*"}),
		MaxAgeSeconds:  aws.Int64(3000),
		AllowedMethods: aws.StringSlice(filterMethods(methods)),
	}

	params := s3.PutBucketCorsInput{
		Bucket: &bucketName,
		CORSConfiguration: &s3.CORSConfiguration{
			CORSRules: []*s3.CORSRule{&rule},
		},
	}

	_, err := s3c.client.PutBucketCors(&params)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to set Bucket %q's CORS, %s", bucketName, err.Error()))
	}

	return nil
}

func filterMethods(methods []string) []string {
	filtered := make([]string, 0, len(methods))
	for _, m := range methods {
		v := strings.ToUpper(m)
		switch v {
		case http.MethodPost, http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodDelete:
			filtered = append(filtered, v)
		}
	}

	return filtered
}

func CreateDefaultBuckets(s3Client *S3Client) (err error) {
	_, alreadyExists, err := s3Client.CreateBucket(BucketPrivateMediaFiles, S3ACLPrivate.ToPtr())
	if err != nil {
		return err
	}
	if !alreadyExists && err == nil {
		err = s3Client.PutBucketCors(BucketPrivateMediaFiles, http.MethodGet)
		if err != nil {
			return err
		}
	}

	_, alreadyExists, err = s3Client.CreateBucket(BucketPublicMediaFiles, S3ACLPublicRead.ToPtr())
	if err != nil {
		return err
	}
	if !alreadyExists && err == nil {
		err = s3Client.PutBucketCors(BucketPublicMediaFiles, http.MethodGet)
		if err != nil {
			return err
		}
	}

	return nil
}
