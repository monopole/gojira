test: generate
	go test ./...

# There's a wee bit of code to generate for enums.
.PHONY: generate
generate:
	go generate ./...

