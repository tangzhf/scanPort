package serverinfo

import (
	"bufio"
	"fmt"
	"io"
	"myssh"
	"os"
	"strconv"
	"strings"
	"time"
)

type DB struct {
	DBName, ip, osUser, osPassword, logPath string
}

func (d *DB) TbsUsage() (htmlstr string, err error) {
	client, err := myssh.NewMyClient(d.ip, d.osUser, d.osPassword)
	if err != nil {
		fmt.Println(" Error => ", err.Error())
		return "", err
	}
	sqlstr := "set linesize 150;\n SELECT a.tablespace_name,total / (1024 * 1024 ) TOTAL, free / (1024 * 1024) FREE_M,(total - free) / (1024 * 1024) USED_M, round((total - free) / total, 4) * 100 \"USED%\" FROM (SELECT tablespace_name, SUM(bytes) free FROM dba_free_space GROUP BY tablespace_name) a,(SELECT tablespace_name, SUM(bytes) total FROM dba_data_files GROUP BY tablespace_name) b WHERE a.tablespace_name = b.tablespace_name;\n exit;"
	outstr, err := myssh.RunShell(client, "echo  '"+sqlstr+"'> tmp.sql&&source .bash_profile&&sqlplus -S / as sysdba @tmp.sql | awk '$0!~/---/&&NR!=1 {print $1,$2,$3,$4,$5}' OFS=\"=\"  ")
	if err != nil {
		fmt.Println(" Error => ", err.Error())
		return "", err
	}

	htmlstr = "<table border=\"1\">"
	for _, val := range strings.Split(outstr, "\n") {
		if val == "====" {

			return htmlstr, nil
		}
		htmlstr = htmlstr + "<tr>"
		for _, v := range strings.Split(val, "=") {
			htmlstr = htmlstr + "<td>" + v + "</td>"
		}
		htmlstr = htmlstr + "</tr>"
	}
	return htmlstr, nil

}

func (d *DB) ParseAlertLog() error {
	fmt.Println(time.Now().Format("20060102-150405") + ": " + d.ip + " server start parse alterlog ")
	client, err := myssh.NewMyClient(d.ip, d.osUser, d.osPassword)
	if err != nil {
		fmt.Println(" Error => ", err.Error())
		return err
	}
	defer client.Close()
	resultstr, err := myssh.RunShell(client, "grep -Eic 'error|ora-' "+d.logPath)

	if err != nil {
		fmt.Println(" Error => ", err.Error())
		return err
	}

	result, err := strconv.Atoi(strings.Trim(resultstr, "\n"))
	if err != nil {
		fmt.Println(" Error => ", err.Error())
		return err
	}
	idx := strings.LastIndex(d.logPath, "/")
	fp := d.logPath[0 : idx+1]
	if result > 0 {

		errlogName := "Err_alertlog" + time.Now().Format("20060102-150405") + ".log"
		_, err := myssh.RunShell(client, "cat "+d.logPath+" >> "+fp+errlogName+" && echo '' > "+d.logPath)
		if err != nil {
			fmt.Println(" Error => ", err.Error())
			return err
		}
		out, err := myssh.RunShell(client, "cat "+fp+errlogName)
		if err != nil {
			fmt.Println(" Error => ", err.Error())
			return err
		}
		// here is  send mail
		myssh.SendMail(d.DBName+" Error log", "plain", out)
	} else {
		_, err := myssh.RunShell(client, "cat "+d.logPath+" >> "+fp+"bak_alterlog.log && echo '' > "+d.logPath)
		if err != nil {
			fmt.Println(" Error => ", err.Error())
			return err
		}
	}
	return nil

}

func (d *DB) Parse(info string) []DB {
	dbs := make([]DB, 0)
	dbsinfo := strings.Split(info, ";")
	for _, value := range dbsinfo {
		vs := strings.Split(value, ",")
		for idx, v := range vs {
			if idx == 0 {
				d.DBName = strings.Split(v, ":")[0]
				x := strings.Split(v, ":")[1]
				d.ip = strings.Split(x, "=")[1]
			}
			switch {
			case strings.Split(v, "=")[0] == "osUser":
				d.osUser = strings.Split(v, "=")[1]
			case strings.Split(v, "=")[0] == "osPassword":
				d.osPassword = strings.Split(v, "=")[1]
			case strings.Split(v, "=")[0] == "logPath":
				d.logPath = strings.Split(v, "=")[1]

			}

		}
		dbs = append(dbs, *d)
	}
	return dbs
}

func (d *DB) InitConf(fpath string) (string, error) {
	strs := ""

	f, err := os.Open(fpath)
	if err != nil {
		return "", err
	}

	defer f.Close()

	br := bufio.NewReader(f)
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF && len(line) < 2 {
			break
		}
		line = strings.TrimSpace(line)
		if len(line) < 2 || line[0] == '#' {
			continue
		}

		if line[:1] == "[" && line[len(line)-1:] == "]" {
			if strs == "" {
				strs = line[1:len(line)-1] + ":"
				continue
			}
			strs = strs[:len(strs)-1] + ";" + line[1:len(line)-1] + ":"
			continue
		}
		strs = strs + line + ","
		if err == io.EOF {
			break
		}
	}
	return strs[:len(strs)-1], nil
}
