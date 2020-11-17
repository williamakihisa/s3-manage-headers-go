1. install go dan git pada os
2. cari path GOPATH dengan command "go env"
3. cari keterangan dengan label berikut : 

set GOPATH=XXXXX\YYYY

4. tempatkan script s3.go, .env dan storylist.json pada path tersebut

5. install package aws : 

go get -u github.com/aws/aws-sdk-go

6. set up config aws credentials di system,
contoh untuk windows :
C:\Users\User\.aws\credential

isi credential s3 tap story: 
[default]
aws_access_key_id = XXXXX
aws_secret_access_key = YYYYY

- Linux/Unix: $HOME/.aws/config
- Windows: %USERPROFILE%\.aws\config


jalankan script dengan command :

go run s3.go

atau bisa dibuild menjadi bash command untuk debian system. dengan command go build s3.go