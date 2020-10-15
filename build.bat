SET GOOS=linux
go build -o bin/stserver cmd/server/main.go
go build -o bin/stclient cmd/client/main.go

SET GOOS=windows
go build -o bin/stserver.exe cmd/server/main.go
go build -o bin/stclient.exe cmd/client/main.go