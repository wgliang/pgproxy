package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/wgliang/pgproxy/proxy"
)

type Client struct {
	db        *sqlx.DB
	timestamp int64
}

// Command line access to pgproxy and provide a friendly display
// interface.
func Command() {
	client := new(Client)
	var err error
	client.db, err = sqlx.Open("postgres", connStr)
	if err != nil {
		glog.Fatalln(err)
	}
	client.timestamp = time.Now().Unix()

	// Set connections num
	client.db.SetMaxIdleConns(5)
	client.db.SetMaxOpenConns(100)

	defer func() {
		client.db.Close()
		if err != nil {
			glog.Errorln(err)
		}
	}()
	fmt.Printf("	pgproxy (%s)\n", VERSION)
	fmt.Println("	Login in:", time.Unix(client.timestamp, 0).Format("2006-01-02 03:04:05 PM"))
	fmt.Println(`	Type "help" for help.`)
	running := true
	reader := bufio.NewReader(os.Stdin)
	for running {
		// Sleep some Nanoseconds wait for event have been deal.
		time.Sleep(300000 * time.Nanosecond)
		fmt.Print("pgproxy#")
		data, _, _ := reader.ReadLine()
		command := string(data)
		if command == "quit" {
			fmt.Println("pgproxy Exit!")
			return
		}
		client.Request(command)
	}
	return
}

// Client request switcher,for different types of sql statement
// calls different requests.
func (c *Client) Request(sql string) {
	index := strings.Index(sql, " ")
	if index == -1 {
		index = len(sql)
	}
	// Choose right function for requests.
	switch strings.ToLower(sql[0:index]) {
	case "select":
		rows, err := c.db.Query(sql)
		if err != nil {
			glog.Errorln(err)
		} else {
			proxy.RowsFormater(rows)
		}
	case "insert", "delete", "update":
		res, err := c.db.Exec(sql)
		if err != nil {
			glog.Errorln(err)
		} else {
			proxy.ResultFormater(res)
		}
	case `\d`, `\l`, `\q`:
		// res := c.db.Exec(sql)
	}
}
