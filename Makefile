all:
	go run main.go api.go utils.go dbUtils.go

test:
	go test ./...
