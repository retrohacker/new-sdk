all: sdk

cli: src/cli/main.go
	go build -o cli src/cli/main.go

packager: src/packager/main.go
	go build -o packager src/packager/main.go

sdk.tar.gz:
	cd data && tar -czf ../sdk.tar.gz ./

sdk: sdk.tar.gz cli packager
	cp cli sdk
	./packager

clean:
	rm -f cli packager sdk.tar.gz sdk

.PHONY: clean
