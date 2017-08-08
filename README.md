# s3sh

>An interactive s3 shell
>This software is using [ishell](https://github.com/abiosoft/ishell) 
>and is distributed under MIT license.

### To build:

- Clone this repository: https://github.com/agilistik/s3sh.git
- go get github.com/abiosoft/ishell
- go get github.com/aws/aws-sdk-go/aws github.com/aws/aws-sdk-go/aws/awserr github.com/aws/aws-sdk-go/service/s3 github.com/aws/aws-sdk-go/service/s3/s3manager github.com/aws/aws-sdk-go/aws/session
- cd s3sh; go build

### To run:
- Make sure you have your credentials saved in ~/.aws/credentials file. 
- On Linux, MacOS, Solaris11:  s3sh [-p profile]
- on Windows:  s3sh.exe [-p profile]
 
Without parameters, the default profile from ~/.aws/credentials will be used.

### Currently supported commands:
<pre>
cd [path]		Change directory.  Without 'path', will change to root '/'
cr [region]		Change region.
desc [name]		Describe 'name', which can be a bucket or an object key.
history			Prints out the history of the commands in the current session.
get [object] [path]	Download the 'object' to 'path'.  If no path specified, download to the current directory.
ls [path]		List contents of the 'path'.  Without a parameter, lists the 'current directory.'
put [object]		Upload 'object' to the current prefix.
pwd			Print current directory (prefix).
</pre>


