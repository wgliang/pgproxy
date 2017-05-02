package proxy

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/wgliang/pgproxy/parser"
)

var (
	testProxyHost  = "127.0.0.1:9090"
	testRemoteHost = "127.0.0.1:5432"
)

func Benchmark_Start(b *testing.B) {
	go Start(testProxyHost, testRemoteHost, parser.GetQueryModificada)
	time.Sleep(3 * time.Second)

	db, err := sqlx.Open("postgres", "host=127.0.0.1 user=postgres password=xxxxx dbname=db port=9090 sslmode=disable")
	if err != nil {
		b.Error(err)
	}
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(100)

	for i := 0; i < b.N; i++ {
		sql := fmt.Sprintf("select id from client where id = %d", i)
		fmt.Println(sql)
		rows, err := db.Query(sql)
		if err != nil {
			b.Error(err)
		} else {
			for rows.Next() {
				var n int
				err = rows.Scan(&n)
				if err != nil {
					b.Error(err)
				} else {
					if n != i {
						b.Errorf("result is not match,n=%d but id=%d", n, i)
					}
				}
			}
		}
	}
	db.Close()
	os.Exit(0)
}

func Test_Start(t *testing.T) {
	go Start(testProxyHost, testRemoteHost, parser.GetQueryModificada)
	time.Sleep(3 * time.Second)

	db, err := sqlx.Open("postgres", "host=127.0.0.1 user=postgres password=xxxxx dbname=db port=9090 sslmode=disable")
	if err != nil {
		t.Error(err)
	}
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(100)

	rows, err := db.Query("select id from client where id = 8 ")
	if err != nil {
		t.Error(err)
	} else {
		for rows.Next() {
			var n int32
			err = rows.Scan(&n)
			if err != nil {
				t.Error(err)
			} else {
				if n != 8 {
					t.Errorf("result is not match,n=%d but id=8", n)
				}
			}
		}
	}
	db.Close()
	os.Exit(0)
}

func Test_getResolvedAddresses(t *testing.T) {
	getResolvedAddresses("127.0.0.1:9090", "127.0.0.1:8080")
}

func Test_getListener(t *testing.T) {
	paddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:9090")
	if err != nil {
		t.Fatal(err)
	}
	getListener(paddr)
}
