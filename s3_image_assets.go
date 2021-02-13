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
    "path/filepath"
    "fmt"
    "strconv"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
    S3_REGION = "ap-southeast-1"
    S3_BUCKET = "webstorieskompas"
    S3_BUCKET_UPLOAD = "webstorieskompas"
    ASSET_PATH = "assets/"
)

func main() {
  http.HandleFunc("/", handler)
  http.ListenAndServe(":9292", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {

    dt := time.Now()
    var new_path = ASSET_PATH+strconv.FormatInt(time.Now().UnixNano() / int64(time.Millisecond), 10)+"/"


    // Create a single AWS session for download and remove
    s, err := session.NewSession(&aws.Config{Region: aws.String(S3_REGION)})
    if err != nil {
        log.Fatal(err)
    }
    i,j,k := 0,0,0
    var newelm []string
    var tempelm []string
    var subList []string
    //list file html from s3
    s3List := handlerList(s, "tap/assets")

    for range s3List {
      if (strings.Contains(strings.ToLower(s3List[i]), "jpg") || strings.Contains(strings.ToLower(s3List[i]), "png") || strings.Contains(strings.ToLower(s3List[i]), "jpeg") || strings.Contains(strings.ToLower(s3List[i]), "webp")){
        var html = strings.Replace(s3List[i], "tap/assets/", "", -1)
        if (len(html) != 0){
           var pos = strings.Index(html, "/")
           var subs = html[0:pos]
           subList = append(subList, subs)
        }
      }
      i++
    }

    tempelm = uniqueArray(subList)

    //check from listing file if no exist then reupload with cache-control
    imageList := readJSONToken("imagelist.json")
    newelm = difference(tempelm, imageList)

    if (len(newelm) > 0){
      fmt.Fprintln(w, "new asset folder found start reuploading and store assets! - "+dt.String())
      fmt.Fprintf(w, "%+q", newelm)
      fmt.Fprintln(w, "%nbsp")

      for range newelm {
         var imageDetail = handlerList(s, "tap/assets/"+newelm[j])
         _, err := os.Stat(new_path+newelm[j])

         if os.IsNotExist(err) {
           errDir := os.MkdirAll(new_path+newelm[j], 0775)
           if errDir != nil {
             log.Fatal(err)
           }
         }

         l := 0
         for range imageDetail {
           var newfilename = strings.Replace(imageDetail[l], "tap/assets/"+newelm[j], "", -1)
           newfilename = strings.Replace(newfilename, "/", "", -1)
           if (len(newfilename) > 0){
           downloadS3(s, imageDetail[l],new_path+newelm[j]+"/"+newfilename)
           fmt.Fprintln(w, "image to download = "+imageDetail[l])
           }
           l++
         }
         imageList = append(imageList, newelm[j])
         j++
      }
      // Create a single AWS session for uploading
      s2, err2 := session.NewSession(&aws.Config{Region: aws.String(S3_REGION)})
      if err2 != nil {
          log.Fatal(err2)
      }
      for range newelm {
         var imageDetail = handlerList(s2, "tap/assets/"+newelm[k])
         y := 0
         for range imageDetail {
           if (strings.Contains(strings.ToLower(imageDetail[y]), "jpg") || strings.Contains(strings.ToLower(imageDetail[y]), "png") || strings.Contains(strings.ToLower(imageDetail[y]), "jpeg") || strings.Contains(strings.ToLower(imageDetail[y]), "webp")){
              fmt.Fprintln(w, imageDetail[y])
              var newfileupload = strings.Replace(imageDetail[y], "tap/assets/"+newelm[k], "", -1)
              newfileupload = strings.Replace(newfileupload, "/", "", -1)
              if (len(newfileupload) > 0){
//                removeS3(imageDetail[y])
                err = AddFileToS3(s2, imageDetail[y], new_path+newelm[k]+"/"+newfileupload)
                if err != nil {
                    log.Fatal(err)
                }else{
                  fmt.Println(newelm[k]+"/"+newfileupload);
 //                 removeFile(newelm[k]+"/"+newfileupload)
                }
              }
           }
           y++
         }
         k++
      }

      writeJSONToken(imageList, "imagelist.json")
      RemoveContents(new_path)
      fin := time.Now()

      fmt.Fprintln(w, "finish process ! - " + fin.String())
    }else{
      finerr := time.Now()
      fmt.Fprintln(w, "no new image asset found! - " + finerr.String())
    }

}

func uniqueArray(element []string) []string {
  encounter := map[string]bool{}
  for v:= range element {
     encounter[element[v]] = true
  }
  result := []string{}
  for key, _ := range encounter {
     result = append(result, key)
  }
  return result
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
  jsonString, _ := json.Marshal(storylist)
  ioutil.WriteFile(filename, jsonString, os.ModePerm)
  // file, _ := os.OpenFile(filename, os.O_CREATE, os.ModePerm)
  // defer file.Close()
  // encoder := json.NewEncoder(file)
  // encoder.Encode(storylist)
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

    if size > 100000 {
      fmt.Println("time to resize this = "+filename)
      size = 100000
    }

    buffer := make([]byte, size)
    file.Read(buffer)

    // panic("stop")

    // Config settings: this is where you choose the bucket, filename, content-type etc.
    // of the file you're uploading.
    _, err = s3.New(s).PutObject(&s3.PutObjectInput{
        Bucket:               aws.String(S3_BUCKET_UPLOAD),
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
  if (checkFile(fileDir)){
   e := os.Remove(fileDir)
   if e != nil {
      log.Fatal(e)
   }
  }
  return
}

func checkFile(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

func RemoveContents(dir string) error {
    d, err := os.Open(dir)
    if err != nil {
        return err
    }
    defer d.Close()
    names, err := d.Readdirnames(-1)
    if err != nil {
        return err
    }
    for _, name := range names {
        err = os.RemoveAll(filepath.Join(dir, name))
        if err != nil {
            return err
        }
    }
    return nil
}
