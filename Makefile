prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep
	if test -d src/github.com/whosonfirst/go-whosonfirst-geojson; then rm -rf src/github.com/whosonfirst/go-whosonfirst-geojson; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-geojson
	cp geojson.go src/github.com/whosonfirst/go-whosonfirst-geojson/geojson.go

deps:   self
	go get -u "github.com/jeffail/gabs"
	go get -u "github.com/dhconnelly/rtreego"
	go get -u "github.com/kellydunn/golang-geo"
fmt:
	go fmt bin/*.go
	go fmt *.go

dump:	self
	go build -o bin/dump bin/dump.go

pip:	self
	go build -o bin/pip bin/pip.go
