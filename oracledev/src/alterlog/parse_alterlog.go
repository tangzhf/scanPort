package main

import (
	"fmt"
	"myssh"
	"serverinfo"
	"time"
	//"os"
	//"strings"
)

func main() {
	db := new(serverinfo.DB)

	dbstrs, err := db.InitConf("server.conf")
	if err != nil {
		fmt.Println(err)
		return
	}

	dbs := db.Parse(dbstrs)
	go GoParseLog(dbs)

	cs := make([]chan string, len(dbs))
	for {
		fmt.Println(time.Now().Hour())
		if time.Now().Hour() == 9 {
			ChanTbsUsage(dbs, cs)
			//fmt.Println(" send file", cs)
		}
		time.Sleep(3600 * time.Second)
	}

}

func GoParseLog(dbs []serverinfo.DB) {
	for {
		for _, d := range dbs {
			d.ParseAlertLog()
		}
		fmt.Println("-----")
		time.Sleep(300 * time.Second)
	}
}

func ChanTbsUsage(dbs []serverinfo.DB, cs []chan string) {
	for idx, d := range dbs {
		cs[idx] = make(chan string, 2048)
		go GoTbsUsage(d, cs[idx])
	}
	htmlstr := ""
	for _, c := range cs {
		s := <-c
		htmlstr = htmlstr + s
		fmt.Println(htmlstr)
	}
	myssh.SendMail("TBS Usage", "html", htmlstr)
}

func GoTbsUsage(db serverinfo.DB, c chan string) {
	htmlstr, err := db.TbsUsage()
	if err != nil {
		fmt.Println(err)
	}
	c <- "<h3>" + db.DBName + " tablespace usage:<h3>" + htmlstr + "</table><br>"

}
