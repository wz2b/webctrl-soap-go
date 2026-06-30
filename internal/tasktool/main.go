package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
)

func main() {
	if len(os.Args) < 2 {
		fatalf("missing task name")
	}

	var err error
	switch os.Args[1] {
	case "build":
		err = build()
	case "publish":
		err = publish()
	case "md-test":
		err = mdTest()
	default:
		err = fmt.Errorf("unknown task %q", os.Args[1])
	}
	if err != nil {
		fatalf("%v", err)
	}
}

func build() error {
	cmds := []string{"trend_export", "areas", "alcsoap_write_value", "search"}
	arches := []string{
		"linux/386",
		"linux/amd64",
		"linux/arm64",
		"darwin/amd64",
		"freebsd/amd64",
		"windows/386",
		"windows/amd64",
		"js/wasm",
	}

	if err := run("go", "mod", "download"); err != nil {
		return err
	}

	for _, commandPackage := range cmds {
		for _, arch := range arches {
			parts := strings.Split(arch, "/")
			if len(parts) != 2 {
				return fmt.Errorf("invalid target %q", arch)
			}
			if err := buildOne(commandPackage, parts[0], parts[1]); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildOne(pkg, goos, goarch string) error {
	source := fmt.Sprintf("./cmd/%s", pkg)
	dest := fmt.Sprintf("%s-%s-%s", pkg, goos, goarch)
	if goos == "windows" {
		dest += ".exe"
	}

	cmd := exec.Command("go", "build", "-v", "-o", dest, source)
	cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func publish() error {
	p := newPublisher()

	for _, name := range []string{
		"trend_export-darwin-amd64",
		"trend_export-linux-386",
		"trend_export-linux-amd64",
		"trend_export-linux-arm64",
		"trend_export-freebsd-amd64",
		"trend_export-windows-386.exe",
		"trend_export-windows-amd64.exe",
	} {
		if err := p.publishArtifact(name); err != nil {
			return err
		}
	}

	return nil
}

type publisher struct {
	uploader *s3manager.Uploader
}

func newPublisher() *publisher {
	sessionConfig := &aws.Config{
		Region:                       aws.String("us-east-1"),
		CredentialsChainVerboseErrors: aws.Bool(true),
	}

	sess := session.Must(session.NewSession(sessionConfig))
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024
		u.LeavePartsOnError = true
	})

	return &publisher{uploader: uploader}
}

func (p *publisher) publishArtifact(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("open %s: %w", name, err)
	}
	defer file.Close()

	input := &s3manager.UploadInput{
		Bucket: aws.String("tools.autofrog.com"),
		Key:    aws.String("artifacts/" + name),
		Body:   file,
	}
	output, err := p.uploader.Upload(input)
	if err != nil {
		return fmt.Errorf("upload %s: %w", name, err)
	}

	fmt.Printf("Published %s\n", output.Location)
	return nil
}

func mdTest() error {
	input, err := os.ReadFile("docs/trend_export.md")
	if err != nil {
		return err
	}

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	output := markdown.ToHTML(input, nil, renderer)
	fmt.Println(string(output))
	return nil
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
