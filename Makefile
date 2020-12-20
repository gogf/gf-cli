
.PHONY: default

default:
	@echo "USAGE: make build/upx"

build:
	gf build main.go
	make upx

upx:
	@upx -9 ./bin/darwin_amd64/*
	@upx -9 ./bin/linux_386/*
	@upx -9 ./bin/linux_amd64/*
