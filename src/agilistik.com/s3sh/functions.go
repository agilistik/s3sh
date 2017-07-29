package main

import (
	"context"
	"os"
	"strings"
	
	"github.com/abiosoft/ishell"

        "github.com/aws/aws-sdk-go/aws"
        "github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
        "github.com/aws/aws-sdk-go/service/s3"
        "github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func cd (c *ishell.Context, pwd *string, service *ServiceSession) {
        d := "/"
        if len(c.Args) > 0 {
                d = c.Args[0]
        }
        if strings.Index(d, "/") == 0 {
                *pwd = d
        } else if d == "." {
        } else if d == ".." {
                if *pwd != "/" {
// Can't slice a string pointer. 
// Until I find a better way, will need to use a temp variable
                        _pwd := *pwd
                        _pwd = _pwd[:len(_pwd) - 2]
                        _pwd = _pwd[:strings.LastIndex(_pwd, "/")]
                        *pwd = _pwd
                }
        } else {
//Check whether the target prefix exists
                list := ls (c, pwd, service)
                updated := false
                for  p,_ := range list {
                        if p ==  d {
                                *pwd = *pwd + d + "/"
                                updated = true
                        }
                }
                if !updated {
                        c.Println("Prefix " + d +  " does not exist.")

                }
        }
        if strings.LastIndex(*pwd, "/") != len(*pwd) - 1{
                *pwd = *pwd + "/"
        }
        if *pwd == "" {
                *pwd = "/"
        }
}


func cr (c *ishell.Context, service *ServiceSession) {
	  if len(c.Args) > 0 {
                                service.Sess = session.Must(session.NewSessionWithOptions(session.Options{
                SharedConfigState: session.SharedConfigEnable,
                Config: aws.Config{Region: &c.Args[0] },
                }))
                                service.Svc = s3.New(service.Sess)
                        }
                        if len(c.Args) == 0 {
                                c.Println("Please specify region name.")
                                return
                        }
}
	

func describe (c *ishell.Context, svc *s3.S3, pwd *string, obj string) {
        bucket := strings.SplitAfter(*pwd, "/")[1]
        if strings.LastIndex(bucket, "/") == len(bucket) -1 {
                 bucket = bucket[:len(bucket) - 1]
        }
        prefix := strings.SplitAfter(*pwd, bucket)[1]
        if strings.Index(prefix, "/") == 0 {
                prefix = prefix[1:]
        }
        if strings.LastIndex(prefix, "/") != len(prefix) - 1 {
                prefix = prefix + "/"
        }
        input := &s3.HeadObjectInput {
                Bucket: aws.String(bucket),
                Key:    aws.String(prefix + obj),
        }

        result, err := svc.HeadObject(input)
        if err != nil {
                if aerr, ok := err.(awserr.Error); ok {
                        switch aerr.Code() {
                                default:
                                        c.Println(aerr.Error())
                        }
                } else {
                        c.Println(err.Error())
                }
        return
        }
        c.Println(result)
}




func get (c *ishell.Context, pwd *string, service *ServiceSession) {
	 key := c.Args[0]
         bucket := strings.SplitAfter(*pwd, "/")[1]
         if strings.LastIndex(bucket, "/") == len(bucket) -1 {
	       	 bucket = bucket[:len(bucket) - 1]
          }
         prefix := strings.SplitAfter(*pwd, bucket)[1]
         if strings.Index(prefix, "/") == 0 {
                 prefix = prefix[1:]
         }
         if strings.LastIndex(prefix, "/") != len(prefix) - 1 {
                 prefix = prefix + "/"
         }
         fullKey := prefix + key
         downloader := s3manager.NewDownloader(service.Sess)
         f, err := os.Create(key)
         if err != nil {
                 c.Println("Can't create local file " + key)
                 c.Println("%v", err)
                 return
         }
         n, err := downloader.Download(f, &s3.GetObjectInput{
                 Bucket: &bucket,
                 Key:  &fullKey,
         })
         if err != nil {
                 c.Printf("Failed to download file %v\n", key)
                 c.Printf("%v\n", err)
                 return
         }
         c.Printf("Downloaded %v bytes\n", n)
}










func ls (c *ishell.Context, pwd *string, service *ServiceSession) map[string]string  {
        list := make(map[string]string)
        if *pwd == "/" {
                result, err := service.Svc.ListBuckets(nil)
                if err != nil {
                        c.Println("Unable to list objects.")
                }
                for _, b := range result.Buckets {
                // find bucket's region
                        ctx := context.Background()
                        region,_ := s3manager.GetBucketRegion (ctx, service.Sess, aws.StringValue(b.Name), "us-west-2")
                        list[aws.StringValue(b.Name)] = aws.StringValue(&region)
                }
        } else {
                // maybe move to main(), and keep path there?
                bucket := strings.SplitAfter(*pwd, "/")[1]
                if strings.LastIndex(bucket, "/") == len(bucket) -1 {
                        bucket = bucket[:len(bucket) - 1]
                }
                c.Println("Bucket: " + bucket)
                prefix := strings.SplitAfter(*pwd, bucket)[1]
                if strings.LastIndex(prefix, "/") == len(prefix) -1 {
                        prefix = prefix[:len(prefix) - 1]
                }
                if strings.Index(prefix, "/") == 0 {
                        prefix = prefix[1:]
                }
                c.Println("Prefix: " + prefix)
                input := &s3.ListObjectsV2Input{
                        Bucket: aws.String(bucket),
                        MaxKeys:aws.Int64(1024),
                        Prefix: &prefix,
                }
                result, err := service.Svc.ListObjectsV2(input)
              if err != nil {
                        if aerr, ok := err.(awserr.Error); ok {
                                switch aerr.Code() {
                                        case s3.ErrCodeNoSuchBucket:
                                                c.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
                                        default:
                                                c.Println(aerr.Error())
                                                }
                                        } else {
                                c.Println(err.Error())
                                }
                        return list
                        }

                keys := NewStrSet()
                for _, item := range result.Contents {
                        key := *item.Key
                        key = key[len(prefix):]
                        if strings.Index(key, "/") == 0 {
                                key = key[1:]
                        }
                        // If key contains a prefix, find the top level:                        
                        if strings.Index(key, "/") > 0 {
                                key = strings.Split(key,"/")[0]
                        }
                        keys.Add(key)
                }

                        for k := range keys.set {
                                list[k] = ""
                        }
        }
        return list
}

func printdir (c *ishell.Context, pwd *string){
        c.Printf("%s\n", *pwd)
}

