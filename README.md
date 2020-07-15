swagger-ui-go
=============

swagger-ui-go provides a Go http handler for hosting [Swagger UI](https://swagger.io/tools/swagger-ui/).

### Usage

#### As a go package
```shell script
go get -u github.com/li-go/swagger-ui-go
```

#### As a command line tool
```shell script
go get -u github.com/li-go/swagger-ui-go/cmd/swagger-ui
# parse a local schema file in YAML
swagger-ui ./openapi.yaml
# parse a local schema file in JSON
swagger-ui ./openapi.json
# parse a remote schema file
swagger-ui https://petstore.swagger.io/v2/swagger.json
# specify a port (default: 8080)
swagger-ui -port 8000 https://petstore.swagger.io/v2/swagger.json
```

### Resources
- [swagger-api/swagger-ui](https://github.com/swagger-api/swagger-ui)
- [swaggo/files](https://github.com/swaggo/files)
