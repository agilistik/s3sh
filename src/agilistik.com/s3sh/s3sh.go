// https://github.com/abiosoft/ishell/blob/master/example/main.go
package main

import (
//	"errors"
//	"fmt"
//	"context"
	"os"
//	"strings"
//	"time"

	"github.com/abiosoft/ishell"

//	"github.com/aws/aws-sdk-go/aws"
//	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
//	"github.com/aws/aws-sdk-go/service/s3/s3manager"

)

func main () {
	pwd := "/"
	var list map [string]string
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		}))
	svc := s3.New(sess)

	shell := ishell.New()

	//display info
	shell.Println("S3 Shell")


	shell.AddCmd(&ishell.Cmd{
		Name: "ls",
		Help: "list objects",
		Func: func(c *ishell.Context){
			list = ls (c, svc, &pwd, sess)
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
			cd(c, svc, &pwd, sess)
		},
	})

	if len(os.Args) > 1 && os.Args[1] == "exit" {
		shell.Process(os.Args[2:]...)
	} else {
		shell.Run()
		shell.Close()
	}
	
}
