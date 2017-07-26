package main

import (
	"context"
	"strings"
	
	"github.com/abiosoft/ishell"

        "github.com/aws/aws-sdk-go/aws"
        "github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
        "github.com/aws/aws-sdk-go/service/s3"
        "github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func cd (c *ishell.Context,svc *s3.S3, pwd *string, sess *session.Session) {
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
                list := ls (c, svc, pwd, sess)
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
                //*pwd = *pwd + d + "/"
        }
        if strings.LastIndex(*pwd, "/") != len(*pwd) - 1{
                *pwd = *pwd + "/"
        }
        if *pwd == "" {
                *pwd = "/"
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


func ls (c *ishell.Context, svc *s3.S3, pwd *string, sess *session.Session) map[string]string  {
//      var list [] string
        list := make(map[string]string)
        if *pwd == "/" {
                //c.Println("Buckets:")
                result, err := svc.ListBuckets(nil)
                if err != nil {
                        c.Println("Unable to list objects.")
                }
                for _, b := range result.Buckets {
                //      c.Printf("* %s created on %s\n", aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
                // find bucket's region
                        ctx := context.Background()
                        //cfg := sess.ClientConfig("s3")
                        region,_ := s3manager.GetBucketRegion (ctx, sess, aws.StringValue(b.Name), "us-west-2")
                        //list = append(list, aws.StringValue(&region) + " " +  aws.StringValue(b.Name))
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
                result, err := svc.ListObjectsV2(input)
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
                        //      c.Println(k)
                                //list = append(list,k)
                                list[k] = ""
                        }
        }
        return list
}

func printdir (c *ishell.Context, pwd *string){
        c.Printf("%s\n", *pwd)
}

