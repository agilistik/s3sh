// https://github.com/abiosoft/ishell/blob/master/example/main.go
package main

import (
//	"errors"
//	"fmt"
//	"context"
	"os"
	"strings"
//	"time"

	"github.com/abiosoft/ishell"

//	"github.com/aws/aws-sdk-go/aws"
//	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

)

type ServiceSession struct {
        Sess *session.Session
        Svc *s3.S3
}


func main () {
	pwd := "/"
	var list map [string]string
//	var uploader s3manager.Uploader
	var downloader *s3manager.Downloader
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		}))
	svc := s3.New(sess)


	shell := ishell.New()
	
	service := ServiceSession {sess, svc}

	//display info
	shell.Println("S3 Shell")

	shell.AddCmd(&ishell.Cmd{
		Name: "get",
		Help: "download an object",
		Func: func(c *ishell.Context){
			key := c.Args[0]
			bucket := strings.SplitAfter(pwd, "/")[1]
			if strings.LastIndex(bucket, "/") == len(bucket) -1 {
                 		bucket = bucket[:len(bucket) - 1]
       			} 
        		prefix := strings.SplitAfter(pwd, bucket)[1]
        		if strings.Index(prefix, "/") == 0 {
                		prefix = prefix[1:]
       			}
        		if strings.LastIndex(prefix, "/") != len(prefix) - 1 {
                		prefix = prefix + "/"
        		}
			fullKey := prefix + key
			downloader = s3manager.NewDownloader(service.Sess)
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
			
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "cr",
		Help: "change region",
		Func: func(c *ishell.Context){
		cr(c, &service)
		},
	})


	shell.AddCmd(&ishell.Cmd{
		Name: "ls",
		Help: "list objects",
		Func: func(c *ishell.Context){
			list = ls (c, &pwd, &service)
			for r,v := range list {
				if len(v) > 0 {	
					c.Print(v + "\t")
				}
				c.Println(r)
			}
		},
	}) 

	shell.AddCmd(&ishell.Cmd{
		Name: "desc",
		Help: "describe object's metadata",
		Func: func (c *ishell.Context) {
			obj := ""
			if len(c.Args) > 0 {
				obj = c.Args[0]
			}
			if len(obj) == 0 {
				c.Println("Please specify object name.")
				return
			}
			
			describe (c, svc, &pwd, obj)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "pwd",
		Help: "print current directory",
		Func: func(c *ishell.Context){
			printdir (c, &pwd)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "cd",
		Help: "change directory",
		Func: func(c *ishell.Context){
			cd(c, &pwd, &service)
		},
	})

	if len(os.Args) > 1 && os.Args[1] == "exit" {
		shell.Process(os.Args[2:]...)
	} else {
		shell.Run()
		shell.Close()
	}
	
}
