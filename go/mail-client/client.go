package main

import (
	"encoding/base64"
	"gopkg.in/gomail.v2"
	"io/ioutil"
	"log"
	"net/smtp"
)


// SMTP 协议过程

// 以163邮箱为例进行说明。在windows下使用telnet程序，
// 远程主机指定为smtp.163.com，端口号指定为25，然后
// 连接smtp.163.com。具体交互如下：

// telnet smtp.163.com 25
// EHLO 163.com
// auth login
// base64加密后的邮箱账号名
// base64加密后的邮箱密码
// MAIL FROM: <XXX@163.com>
// RCPT TO: <XXX@163.com>
// DATA
// from:XXX@163.com
// to:XXX@163.com
// subject:hello,subject!
// Content-Type:multipart/mixed;boundary=a
// Mime-Version:1.0
//
// --a    //注意此处为"--"
// Content-type:text/plain;charset=utf-8
// Content-Transfer-Encoding:quoted-printable
//
// 此处为正文内容
// --a
// Content-type:image/jpg;name=1.jpg
// Content-Transfer-Encoding:base64
//
// 此处为图片的base64编码
// --a--
// .
// quit

// EHLO：客户向对方服务器标识自己的身份；
// auth login：身份认证；
// MAIL FROM：发送者的邮箱地址；
// RCPT TO：标识接收者的邮箱地址；
// DATA：表示邮件的数据部分，可包含图片、正文、附件等。输入完毕以后，以一个"."开始的行作为数据部分的结束；
// quit：退出这次会话，结束邮件发送。

//————————————————
//版权声明：本文为CSDN博主「蜗牛^_^」的原创文章，遵循 CC 4.0 BY-SA 版权协议，转载请附上原文出处链接及本声明。
//原文链接：https://blog.csdn.net/yuyinghua0302/article/details/81062577


type Client struct {}

func (c *Client) SendMailByGomail() {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", "xxxxx@qq.com", "Eiger")
	m.SetHeader("To", "yyyyy@qq.com")
	m.SetHeader("Subject", "Hello，这是邮件主题")
	m.SetBody("text/html", "<h2>这是一份来自SMTP客户端的右键</h2>")
	m.Attach("./mail-client/images/111.png")
	// Attach是作为附件， Embed是嵌入邮件正文

	d := gomail.NewDialer("smtp.qq.com", 587, "374192922@qq.com", "邮箱授权码")
	if err := d.DialAndSend(m); err != nil {
		log.Fatalln(err)
	}
}

func (c *Client) SendMailByNetSmtp() {
	auth := smtp.PlainAuth("", "xxxxx@qq.com", "邮箱授权码", "smtp.qq.com")

	// 要发送的图片附件
	image, _ := ioutil.ReadFile("./mail-client/images/111.png")
	imageBase64 := base64.StdEncoding.EncodeToString(image)

	msg := []byte(
		"from: xxxxx@qq.com\r\n" +
		"to: yyyyy@qq.com\r\n" +
		"Subject: Hello，这是邮件主题\r\n" +
		"Content-Type: multipart/mixed;boundary=a\r\n" +
		"Mime-Version: 1.0\r\n" +
		"\r\n" +
		"--a\r\n" +
		"Content-Type: text/plain;charset=utf-8\r\n" +
		"Content-Transfer-Encoding: quoted-printable\r\n" +
		"\r\n" +
		"这里是邮件正文：这是使用 net/smtp库 构建的SMTP客户端 发来的邮件" +
		"--a\r\n" +
		"Content-Type: image/png;name=111.png\r\n" +
		"Content-Transfer-Encoding: base64\r\n" +
		"\r\n" +
		imageBase64 + "\r\n" +
		"--a\r\n")

	if err := smtp.SendMail("smtp.qq.com:587", auth, "xxxxx@qq.com", []string{"yyyyy@qq.com"}, msg); err != nil {
		log.Fatalln(err)
	}
}

func (c *Client) SendMailByMailSender() {
	sender := NewMailSender(
		"smtp.qq.com:587",
		"xxxxx@qq.com",
		"你的邮箱授权码",
		"xxxxx@qq.com",
		"yyyyy@qq.com",
		"TCP实现SMTP",
		"<h1>基于TCP协议实现简单的SMTP协议</h1>")

		if err := sender.SendMail(); err != nil {
			log.Fatalln(err)
		}
}