
.PHONY: default

default:
	@echo "USAGE: make upx"

upx:
	@upx -9 ./bin/darwin*/*
	@upx -9 ./bin/linux*/*
