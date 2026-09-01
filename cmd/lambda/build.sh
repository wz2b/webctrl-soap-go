GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bootstrap
zip alcSoapWrite.zip bootstrap
