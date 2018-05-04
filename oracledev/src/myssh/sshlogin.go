package myssh

import (
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"golang.org/x/crypto/ssh"
)

// value of contentType is  "plain" or "html"
func SendMail(subject, contentType, body string) {
	auth := smtp.PlainAuth("", "username@xx.com", "password", "smtp.xxx.net")
	nickname := "DBA-zhifang.Tang"
	to := []string{"username@xx.com"}
	user := "username@xx.com"
	content_type := "Content-Type: text/" + contentType + "; charset=UTF-8"
	msg := []byte("To: " + strings.Join(to, ",") + "\r\nFrom: " + nickname +
		"<" + user + ">\r\nSubject: " + subject + "\r\n" + content_type + "\r\n\r\n" + body)
	err := smtp.SendMail("smtp.263.net:25", auth, user, to, msg)
	if err != nil {
		fmt.Printf("send mail error: %v", err)
	}
}

func NewMyClient(ip, osuser, ospwd string) (*ssh.Client, error) {
	myConf := &ssh.ClientConfig{
		User: osuser,
		Auth: []ssh.AuthMethod{
			ssh.Password(ospwd),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	cli, err := ssh.Dial("tcp", ip+":22", myConf)
	if err != nil {
		fmt.Println("open ssh ERR: =>", err)
		return nil, err
	}
	return cli, nil
}

func RunShell(ssh *ssh.Client, shell string) (string, error) {
	se, err := ssh.NewSession()
	if err != nil {
		fmt.Println("ssh newsession ERR: =>", err)
		return "", err
	}
	defer se.Close()
	outbuf, err := se.CombinedOutput(shell)
	if err != nil {
		if string(outbuf) == "0\n" {
			return string(outbuf), nil
		}
		fmt.Println(shell, " shell run ERROR: =>", err)
		return "", err
	}

	//fmt.Println(shell)
	outputstr := string(outbuf)

	return outputstr, nil

}
