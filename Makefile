FLAGS := -trimpath
NOCGO := CGO_ENABLED=0

build:: test-fixture
	go vet && go fmt
	${NOCGO} go build ${FLAGS} -o tempshare

install::
	sudo cp tempshare /usr/local/bin

test-fixture::
ifeq (, $(shell ls ./test-fixture/large))
	mkdir -p ./test-fixture/assets
	# dd if=/dev/zero of=./test-fixture/large bs=1M count=2500
	dd if=/dev/zero of=./test-fixture/smol bs=1M count=25
	dd if=/dev/zero of=./test-fixture/assets/verysmol bs=1M count=1
endif
	@echo ""

run::
	./tempshare ./test-fixture

ci:: test
	echo "done"

test:: build
	bash test.sh

watch::
	ls tempshare.go | entr -rc make build run

build-all:: build
	${NOCGO} GOOS=linux   GOARCH=amd64 go build ${FLAGS} -o builds/tempshare-linux-x64
	${NOCGO} GOOS=linux   GOARCH=arm   go build ${FLAGS} -o builds/tempshare-linux-arm
	${NOCGO} GOOS=linux   GOARCH=arm64 go build ${FLAGS} -o builds/tempshare-linux-arm64
	${NOCGO} GOOS=darwin  GOARCH=amd64 go build ${FLAGS} -o builds/tempshare-mac-x64
	${NOCGO} GOOS=darwin  GOARCH=arm64 go build ${FLAGS} -o builds/tempshare-mac-arm64
	${NOCGO} GOOS=windows GOARCH=amd64 go build ${FLAGS} -o builds/tempshare-windows.exe
	sha256sum builds/* | tee builds/hashes
	echo '```' > builds/buildout
	go version >> builds/buildout
	cat builds/hashes >> builds/buildout
	echo '```' >> builds/buildout

clean::
	rm -f tempshare
	rm -rf builds

