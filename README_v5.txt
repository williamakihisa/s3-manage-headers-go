1. place .aws/credentials according to your system ( unix : ~/.aws/credentials )
2. install golang ( unix : sudo apt-get install golang )
3. grant access modify to /html and /assets ( chmod -R 777 )
4. to generate apps : go build s3.go
5. to create background process : nohup ./s3 > s3log.out & .
6. hit ctrl+D to end session and ignore current process

optionals :

if error occured during build apps ( go build ) check all prequisites services :
1. check GOPATH ( hit go env )
2. define GOPATH ( hit export GOPATH="$HOME/DIRECTORY" )
3. go get "components" eg : "github.com/aws/aws-sdk-go/aws"
