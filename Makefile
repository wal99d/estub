run:
	@go build -o bin/app main.go
	@bin/app
ssh:
	ssh-keygen -f "/Users/Waleedalharthi/.ssh/known_hosts" -R "[localhost]:2222"
	@cat main.go | ssh localhost -p 2222

build:
	@GOOS=linux go build -o bin/app main.go
	@podman build . -t sshy:v1.0.0