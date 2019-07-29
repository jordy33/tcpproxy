package main

import (
	//"encoding/hex"
	"flag"
	"fmt"
//	"log"
	"net"
	"runtime"
    _ "github.com/go-sql-driver/mysql"
    "database/sql"
"strings"
"os"
)

var (
    // DBCon is the connection handle
    // for the database
    db *sql.DB
)

var localAddress *string 
var remoteAddress *string

func writeLog(data string) {
    fields := strings.Split(data,",")
    //fmt.Println("Len fields:",len(fields))
    if len(fields)>1 {
        imei:=fields[1]

       // fmt.Println("Imei:", imei)
        cmd := "INSERT IGNORE INTO imeis (imei) VALUES ('" + imei+ "');";
        //fmt.Println(cmd)
        if db!= nil{
         _, err :=db.Exec(cmd)
         if err != nil {
           //panic(err)
           fmt.Println(err)
         }
        } else {
          fmt.Println("Error db handler lost")
        }
   }
}

func main() {
arg := os.Args[1]
var err error

localAddress = flag.String("l", "64.22.73.112:"+arg, "Local address")
remoteAddress = flag.String("r", "ms03.trackermexico.com.mx:"+arg, "Remote address")

//var localAddress *string = flag.String("l", "64.22.73.112:"+arg, "Local address")
//var remoteAddress *string = flag.String("r", "ms03.trackermexico.com.mx:"+arg, "Remote address")
flag.Parse()
db, err = sql.Open("mysql", "gpscontrol:qazwsxedc@/ms03")
    if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
    }

	fmt.Printf("Listening: %v\nProxying %v\n", *localAddress, *remoteAddress)

	addr, err := net.ResolveTCPAddr("tcp", *localAddress)
	if err != nil {
		panic(err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}

		go proxyConnection(conn)
	}

}

func proxyConnection(conn *net.TCPConn) {
	rAddr, err := net.ResolveTCPAddr("tcp", *remoteAddress)
	//fmt.Println("Go routines:",runtime.NumGoroutine())
	if err != nil {
		panic(err)
	}

	rConn, err := net.DialTCP("tcp", nil, rAddr)
	if err != nil {
		// If the remote is not available
		defer conn.Close()
	}else {
		defer rConn.Close()
		stopchan := make(chan struct{})
		// a channel to signal that it's stopped
		stoppedchan := make(chan struct{})
		// Request loop
		go func() {
			defer close(stoppedchan)
			for {
				data := make([]byte, 1024*1024)
				n, err := conn.Read(data)
				if err != nil {
					//panic(err)
					//fmt.Println("REQUEST Go routines:",runtime.NumGoroutine())
					close(stopchan)
					rConn.Close()
					return
				}else{
					rConn.Write(data[:n])
					//log.Printf("sent:\n%v", hex.Dump(data[:n]))
                                        writeLog(string(data[:n]))
					//fmt.Println(string(data[:n]))
					var mem runtime.MemStats
					runtime.ReadMemStats(&mem)
					//log.Printf("Allocated memory: %fMB. Number of goroutines: %d", float32(mem.Alloc)/1024.0/1024.0, runtime.NumGoroutine())
				}

			}
		}()

		// Response loop
		for {
			data := make([]byte, 1024*1024)
			n, err := rConn.Read(data)
			if err != nil {
				//panic(err)
				//fmt.Println("RESPONSE Go routines:",runtime.NumGoroutine())
				defer conn.Close()
				rConn.Close()
				return
			}else{
				conn.Write(data[:n])
				//log.Printf("received:\n%v", hex.Dump(data[:n]))
				//fmt.Println(string(n))
			}


		}

	}

}

