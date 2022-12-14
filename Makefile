build:
	CGO_ENABLED=0 go build -o bin/shoper_updater cmd/main.go

build-win:
	GOOS=windows GOARCH=amd64 go build -o bin/shoper_updater.exe cmd/main.go

package-win: build-win
	git checkout etc/login.conf data/Stan_mag.txt
	mkdir -p bin/tmp/shoper_updater
	cp -a bin/*.exe windows/start.bat etc/ data/ bin/tmp/shoper_updater/
	cd bin/tmp && zip shoper_updater.zip -r shoper_updater
	mv bin/tmp/shoper_updater.zip bin/
	rm -r bin/tmp
