rm -rf bin
mkdir bin
cp config.sample.json bin
cp rdbak.sh bin

export GOOS=linux
export GOARCH=arm
export GOARM=7
export GOARCHFULL=arm7
export CGO_ENABLED=0
go build -o bin/$GOOS-$GOARCHFULL/rdbak
