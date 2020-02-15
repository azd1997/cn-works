package main

func main() {
	c := &Client{}
	//c.SendMailByGomail()
	//c.SendMailByNetSmtp()
	c.SendMailByMailSender()
}
