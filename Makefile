all: sdk

cli: src/cli/main.go src/cli/cmd/*.go src/cli/logger/*.go
	go build -o cli src/cli/main.go

packager: src/packager/main.go
	go build -o packager src/packager/main.go

sdk.tar.gz: data/
	cd data && tar -czf ../sdk.tar.gz ./

sdk: sdk.tar.gz cli packager
	cp cli sdk
	./packager

clean:
	rm -f cli packager sdk.tar.gz sdk

deps:
	go get github.com/kardianos/osext
	go get github.com/spf13/cobra/cobra
	go get github.com/fatih/color
	go get github.com/alecaivazis/survey
	go get github.com/gosuri/uiprogress

.PHONY: clean deps
