# Momento Demo

```shell
# build
> GOOS=darwin GOARCH=amd64 go build main.go

# run highest error count
> ./main <json_input_file_path> high

# run longest transaction
> ./main <json_input_file_path> long

# test
> go test ./...
```