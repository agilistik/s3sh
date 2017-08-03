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

func cd (c *ishell.Context, pwd *string, service *ServiceSession) string {
        d := "/"
	newPath := "/"
	oneLevelUp := "/"

	var baseDir string
	var pathArr []string

        if len(c.Args) > 0 {
                d = c.Args[0]
		if d == "/" {
			return d
		}
		pathArr = BuildPath(*pwd, d)
		newPath = strings.Join(pathArr, "/")
		if strings.Index(newPath, "//") == 0 {
			newPath = newPath[1:]
		}
		if pathArr[len(pathArr) - 1] != "" {
			baseDir = pathArr[len(pathArr) - 1]
		}
		if len(pathArr) > 2 && pathArr[len(pathArr)  - 2] != "" {
			oneLevelUp = strings.Join(pathArr[:len(pathArr) - 1], "/")
		}
        }
	if d == ".." || d == "../" || len(c.Args) == 0 {
		return newPath
	}
	
	if strings.Index(oneLevelUp, "//") == 0 {
		oneLevelUp = oneLevelUp[1:]
	}


        list,err := _ls (c, &oneLevelUp, service)

	if err == nil {
		for p, _ := range list {
				if p == baseDir {
						return newPath
					}
			}
		c.Println("Path " + d +  " does not exist, or you don't have access to it.")


		}
	return  *pwd
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
	 var currDir string
	 if len(c.Args) > 1 {
// Check whether the directory exists
		_, err := os.Stat(c.Args[1])
		//dirInfo, err := os.Stat(c.Args[1])
		if err != nil {
			err = os.MkdirAll (c.Args[1], 777)
			if err != nil {
				c.Println("The local directory doesn't exist, and can't be created.")
				return
			}
		}
/*
		if ! dirInfo.Mode().isDir() {
			c.Println("The target path exists, and is not a directory.")
			return
			}
*/
		currDir, err = os.Getwd()
		if err != nil {
			c.Println("Can't get the current directory.")
			return
			}
		err = os.Chdir(c.Args[1])
		defer os.Chdir(currDir)
		if err != nil {
			c.Println("Can't change directory.")
			return
		}
				



	}
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


func _ls (c *ishell.Context, pwd *string, service *ServiceSession) (map[string]string, error ) {
	var err error = nil
	list := make(map[string]string)
	target := pwd
	if *target == "/" || *target == "" {
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
                	bucket := strings.SplitAfter(*target, "/")[1]
                	if strings.LastIndex(bucket, "/") == len(bucket) -1 {
                       		 bucket = bucket[:len(bucket) - 1]
               		 }
                	c.Println("Bucket: " + bucket)
                	prefix := strings.SplitAfter(*target, bucket)[1]
                	if strings.LastIndex(prefix, "/") == len(prefix) -1 && len(prefix) > 1 {
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
               		         c.Println("ERROR in ls...")
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
                        return list, err
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
        return list, err
}

func ls (c *ishell.Context, pwd *string, service *ServiceSession) (map[string]string, error)  {
	var err error = nil
	var target *string
	if len(c.Args) > 0 {
		target = &c.Args[0]
		if *target == "." || *target == "./" {
			target = pwd
		} else {
			pathArr := BuildPath(*pwd, *target)
			_target := strings.Join(pathArr, "/")
			if strings.Index(_target, "//") == 0 {
				_target = _target[1:]
			}
			target = &_target
		}
	} else {
		target = pwd
	}
	r,err := _ls  (c,target,service)
	return r, err

}

func printdir (c *ishell.Context, pwd *string){
        c.Printf("%s\n", *pwd)
}

func put (c *ishell.Context, pwd *string, service *ServiceSession) {
	key := c.Args[0]
	// Might need to process the path to get the filename.
	// Consider BuildPath -- and then usng baseName as the key
	// Or, take last element of stirngs.Split(key "/")

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


	
	uploader := s3manager.NewUploader(service.Sess)
	f, err := os.Open(key)
	if err != nil {
		c.Println("Can't open the file.")
		return
	}
	res, err := uploader.Upload(&s3manager.UploadInput{
			Bucket: &bucket,
			Key: 	&fullKey,
			Body:   f,
		})
	c.Printf("File %v uploaded to %v\n", c.Args[0], res.Location)
	//c.Printf("File %v uploaded to %v\n", c.Args[0], aws.StringValue(res.Location))
}
