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
	go get -u code.google.com/p/goauth2/oauth
	go get -u github.com/golang/glog
	go get -u github.com/google/go-github/github
	go get -u github.com/google/go-querystring/query
	go get -u github.com/lib/pq

check:
	go vet ${PKG}/...
	golint ${GOPATH}/src/${PKG}/...

cover:
	go test -cover ${PKG}/...

clean:
	rm -f ./${EXEC}
