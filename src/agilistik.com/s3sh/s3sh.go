package main

import (
//	"errors"
	"fmt"

	"os"

	"github.com/abiosoft/ishell"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type ServiceSession struct {
        Sess *session.Session
        Svc *s3.S3
}


func main () {
	profile := ""
	var sess *session.Session
	switch len(os.Args) {
		case 3:
			if os.Args[1] == "-p" {
				profile = os.Args[2]
				sess = session.Must(session.NewSessionWithOptions(session.Options{
					SharedConfigState: session.SharedConfigEnable,
					Profile: profile,
				}))
			}
		case 1:
				sess = session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
				}))

		default:
			fmt.Printf("Incorrect number of command line arguments: %v\n", len(os.Args) -1)
				os.Exit(1)
	}
	
	pwd := "/"
	var list map [string]string
//	sess := session.Must(session.NewSessionWithOptions(session.Options{
//		SharedConfigState: session.SharedConfigEnable,
//		}))
	svc := s3.New(sess)


	shell := ishell.New()
	
	service := ServiceSession {sess, svc}

	//display info
	shell.Println("S3 Shell")

	shell.AddCmd(&ishell.Cmd{
		Name: "get",
		Help: "download an object",
		Func: func(c *ishell.Context){
			get (c, &pwd, &service)
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
