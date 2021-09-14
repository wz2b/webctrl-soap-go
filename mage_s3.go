//go:build mage
// +build mage

package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"log"
	"os"
)

func Publish() {

	p := newPublisher()

	p.publishArtifact("trend_export-darwin-amd64")
	p.publishArtifact("trend_export-linux-386")
	p.publishArtifact("trend_export-linux-amd64")
	p.publishArtifact("trend_export-linux-arm64")
	p.publishArtifact("trend_export-freebsd-amd64")
	p.publishArtifact("trend_export-windows-386.exe")
	p.publishArtifact("trend_export-windows-amd64.exe")
}

type publisher struct {
	sess     *session.Session
	uploader *s3manager.Uploader
}

func newPublisher() *publisher {
	sessionConfig := &aws.Config{
		Region:                        aws.String("us-east-1"),
		CredentialsChainVerboseErrors: aws.Bool(true),
	}

	sess := session.Must(session.NewSession(sessionConfig))
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024
		u.LeavePartsOnError = true
	})

	return &publisher{
		sess:     sess,
		uploader: uploader,
	}
}

func (this *publisher) publishArtifact(name string) {
	file, err := os.Open(name)
	if err != nil {
		log.Fatalf("Failed to open file %s, %v", name, err)
	}

	input := &s3manager.UploadInput{
		Bucket: aws.String("tools.autofrog.com"),
		Key:    aws.String("artifacts/" + name),
		Body:   file,
	}
	output, err := this.uploader.Upload(input)
	if err != nil {
		log.Fatalf("failed to put file %s, %v", input.Key, err)
		return
	}
	log.Printf("Published %s\n", output.Location)
}
