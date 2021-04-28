
.DEFAULT_GOAL := build

build:
	CGO_CFLAGS="-I/usr/local/opt/openssl/include" CGO_LDFLAGS="-L/usr/local/opt/openssl/lib" go build

buildUbuntu:
	GOOS=linux GOARCH=amd64 CGO_CFLAGS="-I/usr/local/opt/openssl/include" CGO_LDFLAGS="-L/usr/local/opt/openssl/lib" go build

clear:
	rm Thesis_Client