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

// 'cd' to the specified bucket and prefix.
func cd (c *ishell.Context, pwd *string, service *ServiceSession) string {
// The destination path as supplied by the user. Default value:
        d := "/"
// The destination path once processed by the app.  This may be redundant. Default value: 
	newPath := "/"
// The path "one level higher" than the destination:
	oneLevelUp := "/"

// The last element of the destination prefix:
	var baseDir string
// The path, consiting of bucket and prefix, split into an array:
	var pathArr []string
	if len (c.Args) > 1 {
		c.Println("Please provide one argument to 'cd' to, or no arguments to 'cd /'")
		return *pwd
	}	
        if len(c.Args) == 1 {
                d = c.Args[0]
		if d == "/" {
// 'cd' to root.  Nothing else to worry about, just do it
			return d
		}
// 'd' can be an relative or absolute path, and can contain refereces like '.' or '..'
// BuildPath will return an array representing the absolute path to the destination, taking care of the references.
		pathArr = BuildPath(*pwd, d)
// We can use 'd' here. Then, 'newPath' would not be needed.  But this would somewhat complicate "cd .." case: 
		newPath = strings.Join(pathArr, "/")
// The array returned by BuildPath can have more than one empty element in the beginning.  When joined, they produce '//':
		if strings.Index(newPath, "//") == 0 {
			newPath = newPath[1:]
		}
// Get the last element of the prefix.  As long as it's not an empty string; and currently, I don't see a valid case when it could be "":
		if pathArr[len(pathArr) - 1] != "" {
			baseDir = pathArr[len(pathArr) - 1]
		}
// If the path is longer than one level under root, get the path to the parent prefix:
		if len(pathArr) > 2 && pathArr[len(pathArr)  - 2] != "" {
			oneLevelUp = strings.Join(pathArr[:len(pathArr) - 1], "/")
		}
        }
// Here is why we saved 'd' variable:  if we need to move one level up, we can just do it without verifying whether the destination directory exists:
	if d == ".." || d == "../" || len(c.Args) == 0 {
		return newPath
	}
// BasePath and emtpy array elements again...	
	if strings.Index(oneLevelUp, "//") == 0 {
		oneLevelUp = oneLevelUp[1:]
	}

// And now, let's see if the 'baseDir' exists under 'oneLevelUp' -- and remember that newPath = oneLevelUp + / + baseDir
        list,err := _ls (c, &oneLevelUp, service)

	if err == nil {
		for p, _ := range list {
				if p == baseDir {
// baseDir found, newPath is valid:
						return newPath
					}
			}
		c.Println("Path " + d +  " does not exist, or you don't have access to it.")


		}
// If by now we haven't returned a valid new path, then we should just return back whence we came:
	return  *pwd
}

// Change region
func cr (c *ishell.Context, service *ServiceSession) {
	if len(c.Args) != 1 {
		c.Println("Please provide exactly one argument:  the name of the region to change to.")
		return
	} else {
// Need to create a new session
                service.Sess = session.Must(session.NewSessionWithOptions(session.Options{
                SharedConfigState: session.SharedConfigEnable,
                Config: aws.Config{Region: &c.Args[0] },
                }))
                                service.Svc = s3.New(service.Sess)
                        }

 /*                       if len(c.Args) == 0 {
                                c.Println("Please specify region name.")
                                return
                        }
*/
}
	
// Describe a bucket or an object.
// Currently, rather rudimentary. The object should be in the current prefix (pwd)  
// Will neeed to move parameters processing from s3sh.go to this function; and later, add additioal options to get objects metadata.
func describe (c *ishell.Context, svc *s3.S3, pwd *string, obj string) {
// Assuming the first element of the "path" is the bucket...
        bucket := strings.SplitAfter(*pwd, "/")[1]
        if strings.LastIndex(bucket, "/") == len(bucket) -1 {
                 bucket = bucket[:len(bucket) - 1]
        }
// ... and whatever follows, is the prefix. 
        prefix := strings.SplitAfter(*pwd, bucket)[1]
// trim the leading slash
        if strings.Index(prefix, "/") == 0 {
                prefix = prefix[1:]
        }
// and add the trailing slash
        if strings.LastIndex(prefix, "/") != len(prefix) - 1 {
                prefix = prefix + "/"
        }
// finally, get the object's headers:
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

// Download an object from S3.
func get (c *ishell.Context, pwd *string, service *ServiceSession) {
	if len(c.Args) == 0 || len(c.Args) > 2 {
		c.Println("Please provide the key to download, and optionally, the local path to download to.")
		return
	}
	 key := c.Args[0]
// Remember the current directory, to be able to return later.
// In the current implementation, we 'cd' locally to the directory where we want to put the file.
// This serves as an additioanal check that the destinaiton exists, and we shouldn't worry about links, NFS peculiarities and such.
// Consider using full path instead. 

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
// Save the current directory...
		currDir, err = os.Getwd()
		if err != nil {
			c.Println("Can't get the current directory.")
			return
			}
		err = os.Chdir(c.Args[1])
// and don't forget to come back later
		defer os.Chdir(currDir)
		if err != nil {
			c.Println("Can't change directory.")
			return
		}
	}
// Assuming bucket is the first element of the path...
// This is quite similar to 'describe' function -- need to move splitting path into bucket and prefix into a stand-alone function
         bucket := strings.SplitAfter(*pwd, "/")[1]
         if strings.LastIndex(bucket, "/") == len(bucket) -1 {
	       	 bucket = bucket[:len(bucket) - 1]
          }
// What follows bucket in the path, is prefix
         prefix := strings.SplitAfter(*pwd, bucket)[1]
         if strings.Index(prefix, "/") == 0 {
// Remove leading slash...
                 prefix = prefix[1:]
         }
// and add trailing one
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

// List contents of a prefix.
// This is a service function, called only by other functions.  It is not doing processing of command line options.
func _ls (c *ishell.Context, pwd *string, service *ServiceSession) (map[string]string, error ) {
	var err error = nil
// map to return
	list := make(map[string]string)
	target := pwd
// if we don't know what to list (got an epty string), list the root.
// If we need to list the root, we'll be getting buckets, not object keys.
	if *target == "/" || *target == "" {
        	result, err := service.Svc.ListBuckets(nil)
	        if err != nil {
               		c.Println("Unable to list buckets.")
       		}
        	for _, b := range result.Buckets {
        	// find bucket's region
			ctx := context.Background()
               		region,_ := s3manager.GetBucketRegion (ctx, service.Sess, aws.StringValue(b.Name), "us-west-2")
// Add to the map:  "bucketk name" : "region"
               		list[aws.StringValue(b.Name)] = aws.StringValue(&region)
		        }
	        } else {
// Not on the root level, so need to get object keys.
// First, find bucket and prefix -- this is cleary copy-and-paste code, needs to be moved to function.
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
// We've got an error; return whatever we have by now.
                        return list, err
                        }
// Need to make sure we list each 'subdirectory' (next level of prefix) only once:
                keys := NewStrSet()
                for _, item := range result.Contents {
                        key := *item.Key
// cut the prefix off the beginning of the key:
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
// Key, or a part of a prefix, is the key in the map.  Region is the value.  For objects, don't list region -- do it only for the buckets.
// Think of using the value for some other data that might get useful.
                        for k := range keys.set {
                                list[k] = ""
                        }
        }
        return list, err
}

// the 'ls' function that is invoked from s3sh.
func ls (c *ishell.Context, pwd *string, service *ServiceSession) (map[string]string, error)  {
	var err error = nil
	var target *string
	if len(c.Args) > 0 {
		target = &c.Args[0]
// if listing current directory, so be it
		if *target == "." || *target == "./" {
			target = pwd
		} else {
// Build the path, which can be absolute or relative:
			pathArr := BuildPath(*pwd, *target)
// Could not slice a pointer, so need to introduce a temporary variable:
			_target := strings.Join(pathArr, "/")
			if strings.Index(_target, "//") == 0 {
				_target = _target[1:]
			}
			target = &_target
		}
	} else {
// No arguments supplied, so listing the current path:
		target = pwd
	}
	r,err := _ls  (c,target,service)
	return r, err

}

// print the current directory:
func printdir (c *ishell.Context, pwd *string){
        c.Printf("%s\n", *pwd)
}

// upload a file to S3
func put (c *ishell.Context, pwd *string, service *ServiceSession) {
	if len(c.Args) == 0 || len(c.Args) > 2 {
		c.Println("Please specify the local file to be uploaded.") //, and, optionally, the bucket and prefix to upload to.")	
		return
	}
	key := c.Args[0]
	// Might need to process the path to get the filename.
	// Consider BuildPath -- and then usng baseName as the key
	// Or, take last element of stirngs.Split(key "/")
// Getting bucket and prefix -- see comments in other funcs
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
}
