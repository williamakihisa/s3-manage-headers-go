
package main

import (
    "time"
    "strings"
    "bytes"
    "encoding/json"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "fmt"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
    S3_REGION = "YOUR REGION"
    S3_BUCKET = "YOUR BUCKET"
)

func main() {
    dt := time.Now()
    // Create a single AWS session for download and remove
    s, err := session.NewSession(&aws.Config{Region: aws.String(S3_REGION)})
    if err != nil {
        log.Fatal(err)
    }
    i,j,k := 0,0,0
    var newelm []string
    var tempelm []string
    //list file html from s3
    s3List := handlerList(s, "tap/html")
    for range s3List {
      var html = strings.Replace(s3List[i], "tap/html/", "", -1)
      if (len(html) != 0){
        tempelm = append(tempelm, html)
      }
      i++
    }
    //check from listing file if no exist then reupload with cache-control
    storyList := readJSONToken("storylist.json")
    // fmt.Printf("%+q", s3List)
    newelm = difference(tempelm, storyList)
    // fmt.Printf("%+q", diffList)
    // panic("stop")

    if (len(newelm) > 0){
      fmt.Println("new story found start reuploading and store list json! - "+dt.String())
      fmt.Printf("%+q", newelm)
      fmt.Println("%nbsp")
      for range newelm {
         downloadS3(s, "tap/html/"+newelm[j],newelm[j])
         removeS3("tap/html/"+newelm[j])
         storyList = append(storyList, newelm[j])
         j++
      }
      // Create a single AWS session for uploading
      s2, err2 := session.NewSession(&aws.Config{Region: aws.String(S3_REGION)})
      if err2 != nil {
          log.Fatal(err2)
      }
      for range newelm {
         err = AddFileToS3(s2, "tap/html/"+newelm[k], newelm[k])
         if err != nil {
             log.Fatal(err)
         }else{
           removeFile(newelm[k])
         }
         k++
      }
      writeJSONToken(storyList, "storylist.json")
      fin := time.Now()
      fmt.Println("finish process ! - " + fin.String())
    }else{
      finerr := time.Now()
      fmt.Println("no new story found! - " + finerr.String())
    }

}

func difference(a, b []string) []string {
    mb := make(map[string]struct{}, len(b))
    for _, x := range b {
        mb[x] = struct{}{}
    }
    var diff []string
    for _, x := range a {
        if _, found := mb[x]; !found {
            diff = append(diff, x)
        }
    }
    return diff
}

func writeJSONToken(storylist []string, filename string){
  file, _ := os.OpenFile(filename, os.O_CREATE, os.ModePerm)
  defer file.Close()
  encoder := json.NewEncoder(file)
  encoder.Encode(storylist)
}

func downloadS3(sess *session.Session, s3path string, filename string) int {
  bucket := S3_BUCKET
  item := filename

  file, err := os.Create(item)
  if err != nil {
      return 0
  }
  defer file.Close()

  downloader := s3manager.NewDownloader(sess)
  numBytes, err := downloader.Download(file,
      &s3.GetObjectInput{
          Bucket: aws.String(bucket),
          Key:    aws.String(s3path),
      })
  if err != nil {
      return 0
  }
  // fmt.Println("Downloaded", file.Name(), numBytes, "bytes")
  return int(numBytes)
}

func exitErrorf(msg string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, msg+"\n", args...)
    os.Exit(1)
}

func readJSONToken(fileName string) []string {
  var result []string
  jsonFile, err := os.Open(fileName)
  if err != nil {
    fmt.Println(err)
  }
  defer jsonFile.Close()
  byteValue, _ := ioutil.ReadAll(jsonFile)

  json.Unmarshal([]byte(byteValue), &result)

	return result
}


func handlerList(sess *session.Session, fileDir string) []string {
  var output []string
	svc := s3.New(sess)
	res, err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(S3_BUCKET),
    Prefix: aws.String(fileDir),
	})
	if err != nil {
		fmt.Printf("Error listing bucket:\n%v\n", err)
	}
	for _, object := range res.Contents {
		output = append(output, *object.Key)
	}
  return output
}

// AddFileToS3 will upload a single file to S3, it will require a pre-built aws session
// and will set file info like content type and encryption on the uploaded file.
func AddFileToS3(s *session.Session, fileDir string, filename string) error {

    // Open the file for use
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    // Get file size and read the file content into a buffer
    fileInfo, _ := file.Stat()
    var size int64 = fileInfo.Size()
    buffer := make([]byte, size)
    file.Read(buffer)

    // Config settings: this is where you choose the bucket, filename, content-type etc.
    // of the file you're uploading.
    _, err = s3.New(s).PutObject(&s3.PutObjectInput{
        Bucket:               aws.String(S3_BUCKET),
        Key:                  aws.String(fileDir),
        ACL:                  aws.String("private"),
        Body:                 bytes.NewReader(buffer),
        ContentLength:        aws.Int64(size),
        ContentType:          aws.String(http.DetectContentType(buffer)),
        ContentDisposition:   aws.String("attachment"),
        ServerSideEncryption: aws.String("AES256"),
        CacheControl:         aws.String("max-age=31536000, public"),
    })


    return err
}

func removeS3(filename string){
  svc := s3.New(session.New(&aws.Config{
    Region: aws.String(S3_REGION),
  }))

  resp, err := svc.DeleteObject(&s3.DeleteObjectInput{
    Bucket: aws.String(S3_BUCKET),
    Key:    aws.String(filename),
  })
  fmt.Println(resp)
  if err != nil {
    log.Fatal(err)
  }
}

func removeFile(fileDir string){
  e := os.Remove(fileDir)
  if e != nil {
     log.Fatal(e)
  }
  return
}
