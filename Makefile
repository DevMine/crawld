PKG  = github.com/DevMine/crawld
EXEC = crawld

all: check test build

install:
	go install ${PKG}

build:
	go build -o ${EXEC} ${PKG}

test:
	go test -v ${PKG}/...

deps:
	go get -u golang.org/x/oauth2
	go get -u golang.org/x/net/context
	go get -u github.com/golang/glog
	go get -u github.com/google/go-github/github
	go get -u github.com/google/go-querystring/query
	go get -u github.com/lib/pq

dev-deps:
	go get -u github.com/golang/lint/golint

check:
	go vet ${PKG}/...
	golint ${PKG}/...

cover:
	go test -cover ${PKG}/...

clean:
	rm -f ./${EXEC}
