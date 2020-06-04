
test:
	@go test -cover ./...

dist:
	@gox -output "dist/{{.Dir}}_{{.OS}}_{{.Arch}}"

bin: | test
	@gox -osarch="linux/amd64" -output "dist/robo_linux_amd64"

clean:
	rm -fr dist
