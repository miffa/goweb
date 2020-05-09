package sendmail

import (
	//"github.com/go-gomail/gomail"

	"gopkg.in/gomail.v2"
)

type DbmsMail struct {
	dialer     *gomail.Dialer
	authu      string
	authp      string
	stmpserver string
	stmpport   int
}

func (dm *DbmsMail) Init(u, p, addr string, port int) (err error) {
	dm.authu = u
	dm.authp = p
	dm.stmpserver = addr
	dm.stmpport = port
	//d := gomail.NewDialer("smtp.example.com", 587, "user", "123456")
	dm.dialer = gomail.NewDialer(dm.stmpserver, dm.stmpport, dm.authu, dm.authp)
	s, err := dm.dialer.Dial()
	if err != nil {
		return
	}
	s.Close()
	return
}

func (dm DbmsMail) Send(message string, towho []string) (err error) {
	var s gomail.SendCloser
	s, err = dm.dialer.Dial()
	if err != nil {
		return
	}
	defer s.Close()

	m := gomail.NewMessage()
	m.SetHeader("From", "xxx@xxx.com")
	m.SetAddressHeader("To", "nidayede@haha.com", "xiaoming")
	m.SetHeader("Subject", "dbms hdfs notify")
	m.SetBody("text/html", message)
	if err = gomail.Send(s, m); err != nil {
		return
	}
	//m.Reset()
	return

}
