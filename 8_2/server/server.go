// ftp server
package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	"gopl/8_2/ftp"
	server "gopl/8_2/server/ftp"
)

func handleFunc(con net.Conn) {
	defer con.Close()
	fmt.Println("server connect..." + time.Now().String())

	// 身份验证
	// 读取用户名
	var length uint32
	err := binary.Read(con, binary.LittleEndian, &length)
	fmt.Println("getting username..." + time.Now().String())
	if err != nil {
		err = binary.Write(con, binary.LittleEndian, uint32(0))
		if err != nil {
			log.Println(err)
		}
		return
	}
	fmt.Println(length)
	user := make([]byte, length-uint32(binary.Size(length)))
	err = binary.Read(con, binary.LittleEndian, user)
	if err != nil {
		err = binary.Write(con, binary.LittleEndian, uint32(0))
		if err != nil {
			log.Println(err)
		}
		return
	}
	fmt.Println(string(user))

	// 读取密码
	err = binary.Read(con, binary.LittleEndian, &length)
	fmt.Println("getting pwd..." + time.Now().String())
	if err != nil {
		err = binary.Write(con, binary.LittleEndian, uint32(0))
		if err != nil {
			log.Println(err)
		}
		return
	}
	pwd := make([]byte, length-uint32(binary.Size(length)))
	err = binary.Read(con, binary.LittleEndian, pwd)
	if err != nil {
		err = binary.Write(con, binary.LittleEndian, uint32(0))
		if err != nil {
			log.Println(err)
		}
		return
	}

	// 验证用户名密码获取家目录
	fmt.Println("validating..." + time.Now().String())
	validated, cwd := server.Validate(ftp.Sbyte2str(user), ftp.Sbyte2str(pwd))
	if !validated {
		err = binary.Write(con, binary.LittleEndian, uint32(0))
		if err != nil {
			log.Println(err)
		}
		return
	}

	fmt.Println("getting home..." + time.Now().String())
	home := ftp.Str2sbyte(cwd)
	err = binary.Write(con, binary.LittleEndian, uint32(binary.Size(home)))
	if err != nil {
		log.Println(err)
		return
	}
	err = binary.Write(con, binary.LittleEndian, home)
	if err != nil {
		log.Println(err)
		return
	}

	ftpCon := ftp.FtpConn{
		Con:  con,
		Home: cwd,
		Cwd:  cwd,
	}
	ftpServer := server.FtpServer{
		ftpCon,
	}

	// 循环监听命令请求
	fmt.Println("listening..." + time.Now().String())
	for !ftpServer.Exit {
		var length uint32
		err = binary.Read(con, binary.LittleEndian, &length)
		if err != nil {
			log.Println(err)
			return
		}
		var cmdid uint8
		err = binary.Read(con, binary.LittleEndian, &cmdid)
		if err != nil {
			log.Println(err)
			return
		}
		args := make([]byte, length-uint32(binary.Size(cmdid))-uint32(binary.Size(length)))
		err = binary.Read(con, binary.LittleEndian, args)
		if err != nil {
			log.Println(err)
			return
		}

		switch cmdid {
		case ftp.Commands["cd"]:
			err = ftpServer.HandleCd(args)
		case ftp.Commands["ls"]:
			err = ftpServer.HandleLs(args)
		case ftp.Commands["exit"]:
			err = ftpServer.HandleExit(args)
		case ftp.Commands["mkdir"]:
			err = ftpServer.HandleMkdir(args)
		case ftp.Commands["put"]:
			err = ftpServer.HandlePut(args)
		case ftp.Commands["get"]:
			err = ftpServer.HandleGet(args)
		default:
			err = ftpServer.Write([]byte("no command handler."))
		}

		if err != nil {
			log.Println(err)
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", "localhost:5900")
	if err != nil {
		log.Fatal(err)
	}

	for {
		con, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleFunc(con)
	}
}
