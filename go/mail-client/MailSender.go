package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"time"
)

// 基于TCP实现简易的SMTP协议
type MailSender struct {
	// SMTP服务器的用户名，通常为自己的邮箱名
	// 例如qq邮箱，其SMTP服务器为"smtp.qq.com:587"，登录名为邮箱名
	// 个人普通使用时需输入密码才能登陆；第三方客户端登录则需要这个密码对应的授权码
	// 授权码由相应邮件服务提供方生成
	// 网易邮箱password直接输密码，qq邮箱输授权码
	Username string
	// 邮箱用户的密码(授权码)
	Password string
	// SMTP服务器地址
	Host string

	// 发件人邮箱
	From string
	// 收件人邮箱
	To string
	// 邮件主题
	Subject string
	// 邮件正文
	Text string

}


func NewMailSender(host, username, password, from, to, subject, text string) *MailSender {
	return &MailSender{
		Username: username,
		Password: password,
		Host:     host,
		From:     from,
		To:       to,
		Subject:  subject,
		Text:     text,
	}
}

func (s *MailSender) SendMail() error {
	waittime := 5 * time.Second

	// 1. 连接SMTP服务器
	conn, err := net.Dial("tcp", s.Host)
	if err != nil {
		return errors.New(fmt.Sprintf("Dial Host: %v", err))
	}
	defer conn.Close()

	// 2. 等待并读取服务器响应的内容
	if err = s.waitAndRead(conn, waittime); err != nil {
		return errors.New(fmt.Sprintf("Wait Read (after Dial) : %v", err))
	}

	// 3. 回写EHLO消息
	_, _ = conn.Write([]byte("EHLO Juxuny\r\n"))

	// 4. 等待并读取服务器响应的内容
	if err = s.waitAndRead(conn, waittime); err != nil {
		return errors.New(fmt.Sprintf("Wait Read (after EHLO) : %v", err))
	}

	// 5. 回写AUTH LOGIN消息
	_, _ = conn.Write([]byte("AUTH LOGIN\r\n"))

	// 6. 等待并读取服务器响应的内容
	if err = s.waitAndRead(conn, waittime); err != nil {
		return errors.New(fmt.Sprintf("Wait Read (after AUTH LOGIN) : %v", err))
	}

	// 7. 发送base64编码的用户名
	unameBase64 := base64.StdEncoding.EncodeToString([]byte(s.Username))
	fmt.Println("base64 encoded username: ", unameBase64)
	_, _ = conn.Write([]byte(unameBase64 + "\r\n"))

	// 8. 等待并读取服务器响应的内容
	if err = s.waitAndRead(conn, waittime); err != nil {
		return errors.New(fmt.Sprintf("Wait Read (after AUTH username) : %v", err))
	}

	// 9. 发送base64编码的用户密码
	pwdBase64 := base64.StdEncoding.EncodeToString([]byte(s.Password))
	fmt.Println("base64 encoded password: ", pwdBase64)
	_, _ = conn.Write([]byte(pwdBase64 + "\r\n"))

	// 10. 等待并读取服务器响应的内容
	if err = s.waitAndRead(conn, waittime); err != nil {
		return errors.New(fmt.Sprintf("Wait Read (after AUTH password) : %v", err))
	}

	// 11. 发送发件人信息(MAIL FROM消息)
	_, _ = conn.Write([]byte("MAIL FROM: <" + s.From + ">\r\n"))

	// 12. 等待并读取服务器响应的内容
	if err = s.waitAndRead(conn, waittime); err != nil {
		return errors.New(fmt.Sprintf("Wait Read (after MAIL FROM) : %v", err))
	}

	// 13. 发送收件人信息(RCPT TO消息)
	_, _ = conn.Write([]byte("RCPT TO: <" + s.To + ">\r\n"))

	// 14. 等待并读取服务器响应的内容
	if err = s.waitAndRead(conn, waittime); err != nil {
		return errors.New(fmt.Sprintf("Wait Read (after RCPT TO) : %v", err))
	}

	// 15. 发送邮件内容(Header及Body)
	_, _ = conn.Write([]byte("DATA\r\n"))
	_, _ = conn.Write([]byte("From: <" + s.From + ">\r\n"))
	_, _ = conn.Write([]byte("To: <" + s.To + ">\r\n"))
	_, _ = conn.Write([]byte("Subject: <" + s.Subject + ">\r\n"))
	_, _ = conn.Write([]byte("\r\n"))
	_, _ = conn.Write([]byte(s.Text + "\r\n"))
	_, _ = conn.Write([]byte(".\r\n"))	// 注意有个点
	time.Sleep(5e9)	// 睡眠5秒
	_, _ = conn.Write([]byte("QUIT\r\n"))

	// 16. 等待并读取服务器响应的内容
	if err = s.waitAndRead(conn, waittime); err != nil {
		return errors.New(fmt.Sprintf("Wait Read (after DATA) : %v", err))
	}

	return nil
}

func (s *MailSender) waitAndRead(conn net.Conn, waittime time.Duration) error {
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}
	fmt.Println(string(buf[:n]))

	// 另一种读取方法：(在此处经测试是行不通的，i/o超时，奇怪)
	//_ = conn.SetDeadline(time.Now().Add(waittime))
	//data, err := ioutil.ReadAll(conn)
	//if err != nil {
	//	return err
	//}
	//fmt.Println(string(data))

	return nil
}