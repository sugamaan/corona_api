# corona_api

## 環境
go 1.19  
mysql 8.0

## 環境構築

Docker
```shell
# Dockerfileのbuild
$ docker build -t corona_api -f Dockerfile.local . 

# API起動
$ docker compose -f docker-compose-local.yml up api

# APIとMySQL起動
$ docker compose -f docker-compose-local.yml up
```

ホットリロードの導入(air)

```sh
$ go install github.com/cosmtrek/air@latest
$ which air
$ air init
$ cd cmd/
$ air
```

## ローカル
```shell
# ローカルの環境変数を指定して実行
$ GO_ENV=local go run main.go
```

## APIドキュメント
```shell
// アノテーションコメントからAPIドキュメントを更新
$ swag init
$ open http://localhost:8001/swagger/index.html
```