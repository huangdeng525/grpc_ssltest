package main

import (
	"flag"
	"fmt"
	"os"
	"google.golang.org/grpc"
	"github.com/astaxie/beego/logs"
	"bufio"
	"strings"
	"golang.org/x/net/context"
	"sync"
	"github.com/leesumen/tcmd/src/GrpcDcopPb"
	"google.golang.org/grpc/credentials"
	"errors"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
)

type connection struct {
	conn int
	thread int
	mu sync.Mutex
	conns []*grpc.ClientConn
}
var(
	global_M = 0
	crt = "./client.crt"
	key = "./client.key"
	ca  = "./ca.crt"
)

func main() {

	dsn :=flag.String("dsn","127.0.0.1:50051","Enter the server ip address and port,eg:127.0.0.1:8080")
	c :=flag.Int("c",1,"Number of connections")
	t :=flag.Int("t",1,"Number of concurrent")
	m :=flag.Int("s",0,"0 returns the serialization result, 1 returns the result of the string")
	flag.Parse()
	global_M = *m // source value

	if *dsn=="" {
		fmt.Println("Please enter the server ip address and port !")
		os.Exit(0)
	}
	if *c<1 {
		fmt.Println("Param error,c must >=1!")
		os.Exit(0)
	}
	if *t<1 {
		fmt.Println("Param error,t must >=1!!")
		os.Exit(0)
	}
	logs.Info("dsn:[ ",*dsn,"]")
	logs.Info("c [",*c,"]")
	logs.Info("t [",*t,"]")
	logs.Info("s [",*m,"]")
	paramConn :=&connection{conn:*c,thread:*t}
	/**SSL/TSL**/
	// Load the certificates from disk
	certificate, err := tls.LoadX509KeyPair(crt, key)

	if err != nil {
		fmt.Println("could not load server key pair:", err)
		return
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(ca)

	if err != nil {
		fmt.Errorf("could not read ca certificate: %s", err)
	}
	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		errors.New("failed to append client certs")
	}
	// Create the TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		ServerName:   "andisat", // NOTE: this is required!
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	})

	for i:=0 ;i< *c ;i++ {
		go func() {
			conn,err := grpc.Dial(*dsn,grpc.WithTransportCredentials(creds))
			if err != nil {
				logs.Error("Can not to connect the server [",dsn,"]",err)
				os.Exit(0)
			}

			paramConn.mu.Lock()
			paramConn.conns = append(paramConn.conns,conn )
			paramConn.mu.Unlock()
		}()
	}

	fmt.Println("Connect the service success,please enter command...")
	scan := bufio.NewScanner(os.Stdin)
	for(scan.Scan()) {
		line := scan.Text()
		//查看服务器当前配置
		if line == "ifconfig"{
			fmt.Println("Ip address and port [-dsn] :",*dsn)
			fmt.Println("Number of connections [-c] :" ,*c)
			fmt.Println("Number of concurrent [-t] :",*t)
			fmt.Println("Number of concurrent [-s] :",*m)
			continue
		}
		//退出
		if line == "exit" || line == "quit" {
			os.Exit(0)
		}
		line = strings.TrimSpace(line)

		if line !="" {
			go paramConn.exec(line)
			continue
		}else {
			fmt.Println("Please enter command!")
			continue
		}
	}
}
func (conn *connection) exec(line string)  {
	if len(conn.conns)<=0 {
		logs.Error("No connection has been found")
		return
	}
	if conn.thread<=0 {
		logs.Error("No thread")
		return
	}
	for _,c :=range conn.conns { //连接
		for j := 0 ;j<conn.thread ;j++ {

			dcop := GrpcDcopPb.NewGrpcRequestClient(c)
			if global_M == 1 {
				r,err :=dcop.CommandLineResult(context.Background(),&GrpcDcopPb.PbRequest{CmdLine:line})
				if err != nil {
					logs.Error("Could not access service :", err)
					return
				}
				fmt.Println(r.StrResult)
				continue
			}

			r,err := dcop.ProcessRequest(context.Background(),&GrpcDcopPb.PbRequest{CmdLine:line})
			if err != nil {
				logs.Error("Could not access service :", err)
				return
			}


			if r.Result !=0 {
				fmt.Printf("Error code: 0x%o\n",r.Result)
				return
			}
			field :=r.Field

			for i :=0 ;i<len(field) ;i++ {
				fmt.Printf("%+20s|",strings.ToUpper(field[i]))
				if i==len(field )-1{
					fmt.Print("\n")
				}
			}
			row :=r.Row
			for _,col := range row {
				for _,v := range col.ColVal{
					_type :=v.Type
					switch _type {
					case 1,2,3,4,7:fmt.Printf("%20d|",v.DwVal)
					case 5,6:fmt.Printf("%20d|",v.IVal)
					default:
						fmt.Printf("%+20s|",v.StrVal)
					}
				}
				fmt.Print("\n")
			}
			fmt.Println("OK!",r.Count," Records!")
		}
	}
}