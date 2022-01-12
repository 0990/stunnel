SET GOOS=linux
SET GOARCH=amd64
go build -o bin/stserver cmd/server/main.go
go build -o bin/stclient cmd/client/main.go

SET GOOS=windows
SET GOARCH=amd64
go build -o bin/stserver.exe cmd/server/main.go
go build -o bin/stclient.exe cmd/client/main.go

SET GOOS=linux
SET GOARCH=arm64
go build -o bin/stserver_arm cmd/server/main.go
go build -o bin/stclient_arm cmd/client/main.go