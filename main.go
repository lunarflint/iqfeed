package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type sysMsgType int

const (
	currentProtocol sysMsgType = iota
	registerClientAppCompleted
	removeClientAppCompleted
	currentLoginID
	currentPassword
	stats
)

type systemStat struct {
	serverIP               string
	serverPort             int
	maxSymbols             int
	numberOfSymbols        int
	clientsConnected       int
	secondsSinceLastUpdate int
	reconnections          int
	attemptedReconnections int
	startTime              string
	marketTime             string
	status                 string
	iqFeedVersion          string
	loginID                string
	totalKBsRecv           float64
	kbsPerSecRecv          float64
	avgKBsPerSecRecv       float64
	totalKBsSent           float64
	kbsPerSecSent          float64
	avgKBsPerSecSent       float64
}

type sysMsg struct {
	msgType sysMsgType
	auxTxt  string
}

func parseSystemStat(xs []string) (*systemStat, error) {
	serverIP := xs[2]
	serverPort, err := strconv.Atoi(xs[3])
	if err != nil {
		serverPort = 0
	}
	maxSymbols, err := strconv.Atoi(xs[4])
	if err != nil {
		return nil, err
	}
	numberOfSymbols, err := strconv.Atoi(xs[5])
	if err != nil {
		return nil, err
	}
	clientsConnected, err := strconv.Atoi(xs[6])
	if err != nil {
		return nil, err
	}
	secondsSinceLastUpdate, err := strconv.Atoi(xs[7])
	if err != nil {
		return nil, err
	}
	reconnections, err := strconv.Atoi(xs[8])
	if err != nil {
		return nil, err
	}
	attemptedReconnections, err := strconv.Atoi(xs[9])
	if err != nil {
		return nil, err
	}
	startTime := xs[10]
	marketTime := xs[11]
	status := xs[12]
	iqFeedVersion := xs[13]
	loginID := xs[14]
	totalKBsRecv, err := strconv.ParseFloat(xs[15], 64)
	if err != nil {
		return nil, err
	}
	kbsPerSecRecv, err := strconv.ParseFloat(xs[16], 64)
	if err != nil {
		return nil, err
	}
	avgKBsPerSecRecv, err := strconv.ParseFloat(xs[17], 64)
	if err != nil {
		avgKBsPerSecRecv = 0
	}
	totalKBsSent, err := strconv.ParseFloat(xs[18], 64)
	if err != nil {
		return nil, err
	}
	kbsPerSecSent, err := strconv.ParseFloat(xs[19], 64)
	if err != nil {
		return nil, err
	}
	avgKBsPerSecSent, err := strconv.ParseFloat(xs[20], 64)
	if err != nil {
		avgKBsPerSecSent = 0
	}

	s := &systemStat{
		serverIP, serverPort,
		maxSymbols, numberOfSymbols, clientsConnected, secondsSinceLastUpdate,
		reconnections, attemptedReconnections,
		startTime, marketTime, status, iqFeedVersion, loginID,
		totalKBsRecv, kbsPerSecRecv, avgKBsPerSecRecv,
		totalKBsSent, kbsPerSecSent, avgKBsPerSecSent,
	}
	return s, nil
}

func admRecv(admConn net.Conn, sysMsgCh chan *sysMsg, statMsgCh chan *systemStat) {
	reader := bufio.NewReader(admConn)

	for {
		str, err := reader.ReadString('\n')

		if err != nil {
			log.Fatal(err)
		}

		xs := strings.Split(str, ",")

		switch xs[1] {
		case "CURRENT PROTOCOL":
			sysMsgCh <- &sysMsg{currentProtocol, xs[2]}
		case "REGISTER CLIENT APP COMPLETED":
			sysMsgCh <- &sysMsg{registerClientAppCompleted, ""}
		case "REMOVE CLIENT APP COMPLETED":
			sysMsgCh <- &sysMsg{removeClientAppCompleted, ""}
		case "CURRENT LOGINID":
			sysMsgCh <- &sysMsg{currentLoginID, xs[2]}
		case "CURRENT PASSWORD":
			sysMsgCh <- &sysMsg{currentPassword, xs[2]}
		case "STATS":
			s, err := parseSystemStat(xs)
			if err == nil {
				statMsgCh <- s
			} else {
				log.Println(str)
			}
		default:
			log.Println(str)
		}
	}
}

//type adminConn net.Conn

func setProtocolCmd(version string) []byte {
	str := fmt.Sprintf("S,SET PROTOCOL,%s\r\n", version)
	return []byte(str)
}

func setClientNameCmd(name string) []byte {
	str := fmt.Sprintf("S,SET CLIENT NAME,%s\r\n", name)
	return []byte(str)
}

func registerClientAppCmd(productID string, productVersion string) []byte {
	str := fmt.Sprintf("S,REGISTER CLIENT APP,%s,%s\r\n", productID, productVersion)
	return []byte(str)
}

func removeClientAppCmd(productID string, productVersion string) []byte {
	str := fmt.Sprintf("S,REMOVE CLIENT APP,%s,%s\r\n", productID, productVersion)
	return []byte(str)
}

func setLoginIDCmd(id string) []byte {
	str := fmt.Sprintf("S,SET LOGINID,%s\r\n", id)
	return []byte(str)
}

func setPasswordCmd(pw string) []byte {
	str := fmt.Sprintf("S,SET PASSWORD,%s\r\n", pw)
	return []byte(str)
}

func connectCmd() []byte {
	str := "S,CONNECT\r\n"
	return []byte(str)
}

func disconnectCmd() []byte {
	str := "S,DISCONNECT\r\n"
	return []byte(str)
}

func main() {
	protocolVersion := "6.1"
	ptclVer := os.Getenv("IQFEED_PTCLVER")
	if ptclVer != "" {
		protocolVersion = ptclVer
	}

	loginID := os.Getenv("IQFEED_LOGINID")
	password := os.Getenv("IQFEED_PASSWD")
	productID := os.Getenv("IQFEED_PRODID")
	productVersion := os.Getenv("IQFEED_PRODVER")

	shouldProxy := os.Getenv("IQFEED_PROXY") == "YES"

	cmd := exec.Command("xvfb-run", "-s", "-noreset", "-a", "wine", "iqconnect.exe")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	sysMsgCh := make(chan *sysMsg, 5)
	systemStatCh := make(chan *systemStat, 10)

	var admConn net.Conn

	for i := 0; i < 50; i++ {
		admConn, err = net.Dial("tcp", fmt.Sprintf("localhost:%d", 9300))

		if err == nil {
			break
		} else {
			time.Sleep(200 * time.Millisecond)
		}
	}

	if err != nil {
		log.Fatal(err)
	}

	go admRecv(admConn, sysMsgCh, systemStatCh)

	srcHostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	dstHostname := "localhost"

	if shouldProxy {
		go proxy(srcHostname, 5009, dstHostname, 5009)
		go proxy(srcHostname, 9100, dstHostname, 9100)
		go proxy(srcHostname, 9200, dstHostname, 9200)
		go proxy(srcHostname, 9300, dstHostname, 9300)
		go proxy(srcHostname, 9400, dstHostname, 9400)
	}

	var sysMsg *sysMsg

	admConn.Write(setProtocolCmd(protocolVersion))
	sysMsg = <-sysMsgCh
	if sysMsg.msgType != currentProtocol || sysMsg.auxTxt != protocolVersion {
		log.Fatal("Unexpected response")
	}

	admConn.Write(setClientNameCmd("ADMIN_CONN"))

	admConn.Write(registerClientAppCmd(productID, productVersion))
	sysMsg = <-sysMsgCh
	if sysMsg.msgType != registerClientAppCompleted {
		log.Fatal("Unexpected response")
	}

	admConn.Write(setLoginIDCmd(loginID))
	sysMsg = <-sysMsgCh
	if sysMsg.msgType != currentLoginID || sysMsg.auxTxt != loginID {
		log.Fatal("Unexpected response")
	}

	admConn.Write(setPasswordCmd(password))
	sysMsg = <-sysMsgCh
	if sysMsg.msgType != currentPassword {
		log.Fatal("Unexpected response")
	}

	admConn.Write(connectCmd())

	lastMarketTime := ""
	for {
		select {
		case msg := <-sysMsgCh:
			log.Println(msg)

		case stat := <-systemStatCh:
			if stat == nil {
				log.Fatal("Unexpected response")
			}

			ox := "O"
			if stat.status != "Connected" {
				ox = "X"
			}
			if lastMarketTime != stat.marketTime {
				log.Printf("[Status] Mkt time: %s Login ID: %s Client start time: %s\n", stat.marketTime, stat.loginID, stat.startTime)
				lastMarketTime = stat.marketTime
			}

			log.Printf("[%s] Send: %.2f kB/s Recv: %.2f kB/s Sym: %d/%d #Conn: %d Reconn: %d/%d\n", ox, stat.kbsPerSecRecv, stat.kbsPerSecSent, stat.numberOfSymbols, stat.maxSymbols, stat.clientsConnected, stat.reconnections, stat.attemptedReconnections)
		}
	}
}
