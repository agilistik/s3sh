package main

import (
//	"errors"
	"fmt"

	"os"
	"sort"
	"strings"

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
	histSize := 50
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

	h := NewHist(histSize)	
	pwd := "/"
	var list map [string]string
	svc := s3.New(sess)


	shell := ishell.New()
	
	service := ServiceSession {sess, svc}

	//display info
	shell.Println("S3 Shell")

	shell.AddCmd(&ishell.Cmd{
		Name: "put",
		Help: "upload a file to s3",
		Func:  func(c *ishell.Context){
			h.Add("put " + strings.Join(c.Args, " "))
			put (c, &pwd, &service)
		},
	})



	shell.AddCmd(&ishell.Cmd{
		Name: "get",
		Help: "download an object",
		Func: func(c *ishell.Context){
			h.Add("get " + strings.Join(c.Args, " "))
			get (c, &pwd, &service)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "cr",
		Help: "change region",
		Func: func(c *ishell.Context){
		h.Add("cr " + strings.Join(c.Args, " "))
		cr(c, &service)
		},
	})


	shell.AddCmd(&ishell.Cmd{
		Name: "history",
		Help: "show history of commands",
		Func: func(c *ishell.Context){
		h.Add("history")
		history(c, h)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "ls",
		Help: "list objects",
		Func: func(c *ishell.Context){
			h.Add("ls " + strings.Join(c.Args, " "))
// Get the map with the 'ls' results...
			list,_ = ls (c, &pwd, &service)
// build an array of the keys in the results map...
			keys := make([]string, len(list))
			for r:= range list {
				keys = append(keys, r)
			}
// sort the array
			sort.Strings(keys)
// and print the results, sorted.
			for r := range keys {
				if len(list[keys[r]]) > 0 {	
					c.Print(list[keys[r]] + "\t")
				}
				if (len(keys[r]) > 0) {
					c.Println(keys[r])
				}
			}
		},
	}) 

	shell.AddCmd(&ishell.Cmd{
		Name: "desc",
		Help: "describe object's metadata",
		Func: func (c *ishell.Context) {
			h.Add("desc " + strings.Join(c.Args, " "))
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
			h.Add("pwd " + strings.Join(c.Args, " "))
			printdir (c, &pwd)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "cd",
		Help: "change directory",
		Func: func(c *ishell.Context){
		h.Add("cd " + strings.Join(c.Args, " "))
		pwd =	cd(c, &pwd, &service)
		},
	})

	if len(os.Args) > 1 && os.Args[1] == "exit" {
		shell.Process(os.Args[2:]...)
	} else {
		shell.Run()
		shell.Close()
	}
	
}
