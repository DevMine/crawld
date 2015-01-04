PKG  = github.com/DevMine/crawld
EXEC = crawld

all: check test build

install:
	go install ${PKG}

build:
	go build -o ${EXEC} ${PKG}

test:
	go test ${PKG}/...

deps:
	 go get -u github.com/DevMine/crawld

check:
	go vet ${PKG}/...
	golint ${GOPATH}/src/${PKG}

cover:
	go test -cover ${PKG}/...

clean:
	rm -f ./${EXEC}
