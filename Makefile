build:
	go build -ldflags="-s -w" -o gcr main.go && upx -9 gcr && chmod +x gcr
gcr: build
install: gcr
	sudo cp -fr gcr /usr/local/bin/gcr