This project provides a **Go-based PubChem API library** designed to simplify the process of obtaining compound information from a PubChem database. With this library, developers can easily query and extract various attribute data of compounds.

Using a different logic from PubChempy to retrieve compound information, this approach allows for more diverse and precise queries.

The query results are stored in a database for convenient local use in the future.
## Database config
default: databases based on MySQL protocol
```shell
# modify this config
var MysqlCursor = getMysqlCursor("192.168.2.139", 2881, "luocx@aidb", "ABab12@#", "enotess")
```

## Run
```shell
# clone repository
git clone https://github.com/cx-luo/go-pubchem.git

cd go-pubchem

# run
go run main.go

## or
# build & run
go mod tidy
go build -o ../goPubChem
../goPubChem
```

**Swagger url:** 

http://127.0.0.1:8100/swagger/index.html