package pack

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

var (
	negotiateProtocolRequest, _  = hex.DecodeString("00000085ff534d4272000000001853c00000000000000000000000000000fffe00004000006200025043204e4554574f524b2050524f4752414d20312e3000024c414e4d414e312e30000257696e646f777320666f7220576f726b67726f75707320332e316100024c4d312e325830303200024c414e4d414e322e3100024e54204c4d20302e313200")
	sessionSetupRequest, _       = hex.DecodeString("00000088ff534d4273000000001807c00000000000000000000000000000fffe000040000dff00880004110a000000000000000100000000000000d40000004b000000000000570069006e0064006f007700730020003200300030003000200032003100390035000000570069006e0064006f007700730020003200300030003000200035002e0030000000")
	treeConnectRequest, _        = hex.DecodeString("00000060ff534d4275000000001807c00000000000000000000000000000fffe0008400004ff006000080001003500005c005c003100390032002e003100360038002e003100370035002e003100320038005c00490050004300240000003f3f3f3f3f00")
	transNamedPipeRequest, _     = hex.DecodeString("0000004aff534d42250000000018012800000000000000000000000000088ea3010852981000000000ffffffff0000000000000000000000004a0000004a0002002300000007005c504950455c00")
	trans2SessionSetupRequest, _ = hex.DecodeString("0000004eff534d4232000000001807c00000000000000000000000000008fffe000841000f0c0000000100000000000000a6d9a40000000c00420000004e0001000e000d0000000000000000000000000000")
)

func detectHost(ip string) {
	// Connecting to a host in LAN, if reachable, should be very quick
	timeout := time.Second * 2
	conn, err := net.DialTimeout("tcp", ip+":445", timeout)
	if err != nil {
		fmt.Printf("Failed to connect to %s: %v\n", ip, err)
		return
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(time.Second * 5))
	conn.Write(negotiateProtocolRequest)
	reply := make([]byte, 1024)

	// Let alone half packet
	if n, err := conn.Read(reply); err != nil || n < 36 {
		return
	}

	if binary.LittleEndian.Uint32(reply[9:13]) != 0 {
		// Received error
		return
	}

	conn.Write(sessionSetupRequest)

	n, err := conn.Read(reply)
	if err != nil || n < 36 {
		return
	}

	if binary.LittleEndian.Uint32(reply[9:13]) != 0 {
		// Received error
		fmt.Printf("Can't determine whether %s is vulnerable or not\n", ip)
		return
	}

	// Extract OS info
	var os string
	sessionSetupResponse := reply[36:n]
	if wordCount := sessionSetupResponse[0]; wordCount != 0 {
		// Find byte count
		byteCount := binary.LittleEndian.Uint16(sessionSetupResponse[7:9])
		if n != int(byteCount)+45 {
			fmt.Println("Invalid session setup AndX response")
		} else {
			// Two continuous null bytes as the end of a Unicode string
			for i := 10; i < len(sessionSetupResponse)-1; i++ {
				if sessionSetupResponse[i] == 0 && sessionSetupResponse[i+1] == 0 {
					os = string(sessionSetupResponse[10:i])
					break
				}
			}
		}
	}

	userID := reply[32:34]
	treeConnectRequest[32] = userID[0]
	treeConnectRequest[33] = userID[1]

	// TODO: Change the IP in the tree path, though it doesn't matter
	conn.Write(treeConnectRequest)

	if n, err := conn.Read(reply); err != nil || n < 36 {
		return
	}

	treeID := reply[28:30]
	transNamedPipeRequest[28] = treeID[0]
	transNamedPipeRequest[29] = treeID[1]
	transNamedPipeRequest[32] = userID[0]
	transNamedPipeRequest[33] = userID[1]

	conn.Write(transNamedPipeRequest)
	if n, err := conn.Read(reply); err != nil || n < 36 {
		return
	}

	if reply[9] == 0x05 && reply[10] == 0x02 && reply[11] == 0x00 && reply[12] == 0xc0 {
		fmt.Printf("%s(%s) is likely VULNERABLE to MS17-010!\n", ip, os)

		// Detect the presence of DOUBLEPULSAR SMB implant
		trans2SessionSetupRequest[28] = treeID[0]
		trans2SessionSetupRequest[29] = treeID[1]
		trans2SessionSetupRequest[32] = userID[0]
		trans2SessionSetupRequest[33] = userID[1]

		conn.Write(trans2SessionSetupRequest)

		if n, err := conn.Read(reply); err != nil || n < 36 {
			return
		}

		if reply[34] == 0x51 {
			fmt.Printf("DOUBLEPULSAR SMB IMPLANT in %s\n", ip)
			writeToVulnerableFile(ip)
		}
	} else {
		fmt.Printf("%s(%s) stays in safety\n", ip, os)
	}
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func readHostsFromFile(filePath string) ([]string, error) {
	var hosts []string

	file, err := os.Open(filePath)
	if err != nil {
		return hosts, fmt.Errorf("error opening %s: %v", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		hosts = append(hosts, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return hosts, fmt.Errorf("error reading %s: %v", filePath, err)
	}

	return hosts, nil
}

func getLocalCIDR() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("error getting interface addresses: %v", err)
	}

	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err == nil && ip.To4() != nil {
			return ip.String() + "/24", nil
		}
	}

	return "", fmt.Errorf("unable to determine local CIDR")
}

func writeToVulnerableFile(ip string) {
	filePath := "C:\\Windows\\loveorhate\\Vulnerable.txt"
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening or creating", filePath, ":", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(ip + "\n")
	if err != nil {
		fmt.Println("Error writing to", filePath, ":", err)
	}
}


func Vunlscanner() {
	// Read SMB hosts from smbip.txt
	hosts, err := readHostsFromFile("smbip.txt")
	if err != nil {
		fmt.Println("Error reading smbip.txt:", err)
		return
	}

	// Get local CIDR
	cidr, err := getLocalCIDR()
	if err != nil {
		fmt.Println("Error getting local CIDR:", err)
		return
	}

	var wg sync.WaitGroup

	// Scan SMB hosts
	for _, h := range hosts {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			detectHost(ip)
		}(h)
	}

	// Scan local network
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		fmt.Println("Error parsing local CIDR:", err)
		return
	}

	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); incIP(ip) {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			detectHost(ip)
		}(ip.String())
	}

	wg.Wait()
}
