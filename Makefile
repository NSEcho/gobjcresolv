# in the case you are having brew installed,
# make sure that ranlib --version does not show GNU
# if it does, make sure that /usr/lib is
# before /usr/local/lib inside $PATH
CGO_ENABLED = 1
GOOS = ios
GOARCH = arm64
OUTPUT = gobjcresolv.dylib
CC = $(shell xcrun --sdk iphoneos --find clang) \
-arch arm64 \
-isysroot $(shell xcrun --sdk iphoneos --show-sdk-path)
all:
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) \
		GOARCH=$(GOARCH) CC="$(CC)" \
		go build -buildmode=c-archive -o gobjcresolv.a .
	@xcrun --sdk iphoneos clang -arch arm64 \
		-shared -all_load -o $(OUTPUT) gobjcresolv.a \
		-framework CoreFoundation
	@rm gobjcresolv.h gobjcresolv.a
	@echo "[*] Created dylib $(OUTPUT)"
