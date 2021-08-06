package ip2region

import (
	"bytes"
	"embed"
	"errors"
	"strconv"
	"strings"
)

const (
	INDEX_BLOCK_LENGTH  = 12
	TOTAL_HEADER_LENGTH = 8192
)

var err error
var ipInfo IpInfo

//go:embed ip2region.db
var f embed.FS

type Ip2Region struct {
	// db file handler
	dbFileHandler *bytes.Reader

	//header block info

	headerSip []int64
	headerPtr []int64
	headerLen int64

	// super block index info
	firstIndexPtr int64
	lastIndexPtr  int64
	totalBlocks   int64

	// for memory mode only
	// the original db binary string

	dbBinStr []byte
}

type IpInfo struct {
	CityId   int64
	Country  string
	Region   string
	Province string
	City     string
	ISP      string
}

func (ip IpInfo) String() string {
	return strconv.FormatInt(ip.CityId, 10) + "|" + ip.Country + "|" + ip.Region + "|" + ip.Province + "|" + ip.City + "|" + ip.ISP
}

func getIpInfo(cityId int64, line []byte) IpInfo {

	lineSlice := strings.Split(string(line), "|")
	ipInfo := IpInfo{}
	length := len(lineSlice)
	ipInfo.CityId = cityId
	if length < 5 {
		for i := 0; i <= 5-length; i++ {
			lineSlice = append(lineSlice, "")
		}
	}

	ipInfo.Country = lineSlice[0]
	ipInfo.Region = lineSlice[1]
	ipInfo.Province = lineSlice[2]
	ipInfo.City = lineSlice[3]
	ipInfo.ISP = lineSlice[4]
	return ipInfo
}

func New() (*Ip2Region, error) {
	bys, err := f.ReadFile("ip2region.db")
	if err != nil {
		return nil, err
	}
	return &Ip2Region{
		dbBinStr:      bys,
		dbFileHandler: bytes.NewReader(bys),
	}, nil
}

func (t *Ip2Region) Close() {
	//t.dbFileHandler.Close()
}

func (t *Ip2Region) MemorySearch(ipStr string) (ipInfo IpInfo, err error) {
	ipInfo = IpInfo{}
	if t.totalBlocks == 0 {
		t.firstIndexPtr = getLong(t.dbBinStr, 0)
		t.lastIndexPtr = getLong(t.dbBinStr, 4)
		t.totalBlocks = (t.lastIndexPtr-t.firstIndexPtr)/INDEX_BLOCK_LENGTH + 1
	}

	ip, err := ip2long(ipStr)
	if err != nil {
		return ipInfo, err
	}

	h := t.totalBlocks
	var dataPtr, l int64
	for l <= h {

		m := (l + h) >> 1
		p := t.firstIndexPtr + m*INDEX_BLOCK_LENGTH
		sip := getLong(t.dbBinStr, p)
		if ip < sip {
			h = m - 1
		} else {
			eip := getLong(t.dbBinStr, p+4)
			if ip > eip {
				l = m + 1
			} else {
				dataPtr = getLong(t.dbBinStr, p+8)
				break
			}
		}
	}
	if dataPtr == 0 {
		return ipInfo, errors.New("not found")
	}

	dataLen := (dataPtr >> 24) & 0xFF
	dataPtr = dataPtr & 0x00FFFFFF
	ipInfo = getIpInfo(getLong(t.dbBinStr, dataPtr), t.dbBinStr[(dataPtr)+4:dataPtr+dataLen])
	return ipInfo, nil

}

func (t *Ip2Region) BinarySearch(ipStr string) (ipInfo IpInfo, err error) {
	ipInfo = IpInfo{}
	if t.totalBlocks == 0 {
		t.dbFileHandler.Seek(0, 0)
		superBlock := make([]byte, 8)
		t.dbFileHandler.Read(superBlock)
		t.firstIndexPtr = getLong(superBlock, 0)
		t.lastIndexPtr = getLong(superBlock, 4)
		t.totalBlocks = (t.lastIndexPtr-t.firstIndexPtr)/INDEX_BLOCK_LENGTH + 1
	}

	var l, dataPtr, p int64

	h := t.totalBlocks

	ip, err := ip2long(ipStr)

	if err != nil {
		return
	}

	for l <= h {
		m := (l + h) >> 1

		p = m * INDEX_BLOCK_LENGTH

		_, err = t.dbFileHandler.Seek(t.firstIndexPtr+p, 0)
		if err != nil {
			return
		}

		buffer := make([]byte, INDEX_BLOCK_LENGTH)
		_, err = t.dbFileHandler.Read(buffer)

		if err != nil {

		}
		sip := getLong(buffer, 0)
		if ip < sip {
			h = m - 1
		} else {
			eip := getLong(buffer, 4)
			if ip > eip {
				l = m + 1
			} else {
				dataPtr = getLong(buffer, 8)
				break
			}
		}

	}

	if dataPtr == 0 {
		err = errors.New("not found")
		return
	}

	dataLen := (dataPtr >> 24) & 0xFF
	dataPtr = dataPtr & 0x00FFFFFF

	t.dbFileHandler.Seek(dataPtr, 0)
	data := make([]byte, dataLen)
	t.dbFileHandler.Read(data)
	ipInfo = getIpInfo(getLong(data, 0), data[4:dataLen])
	err = nil
	return
}

func (t *Ip2Region) BtreeSearch(ipStr string) (ipInfo IpInfo, err error) {
	ipInfo = IpInfo{}
	ip, err := ip2long(ipStr)

	if t.headerLen == 0 {
		t.dbFileHandler.Seek(8, 0)

		buffer := make([]byte, TOTAL_HEADER_LENGTH)
		t.dbFileHandler.Read(buffer)
		var idx int64
		for i := 0; i < TOTAL_HEADER_LENGTH; i += 8 {
			startIp := getLong(buffer, int64(i))
			dataPar := getLong(buffer, int64(i+4))
			if dataPar == 0 {
				break
			}

			t.headerSip = append(t.headerSip, startIp)
			t.headerPtr = append(t.headerPtr, dataPar)
			idx++
		}

		t.headerLen = idx
	}

	var l, sptr, eptr int64
	h := t.headerLen

	for l <= h {
		m := (l + h) >> 1
		if m < t.headerLen {
			if ip == t.headerSip[m] {
				if m > 0 {
					sptr = t.headerPtr[m-1]
					eptr = t.headerPtr[m]
				} else {
					sptr = t.headerPtr[m]
					eptr = t.headerPtr[m+1]
				}
				break
			}
			if ip < t.headerSip[m] {
				if m == 0 {
					sptr = t.headerPtr[m]
					eptr = t.headerPtr[m+1]
					break
				} else if ip > t.headerSip[m-1] {
					sptr = t.headerPtr[m-1]
					eptr = t.headerPtr[m]
					break
				}
				h = m - 1
			} else {
				if m == t.headerLen-1 {
					sptr = t.headerPtr[m-1]
					eptr = t.headerPtr[m]
					break
				} else if ip <= t.headerSip[m+1] {
					sptr = t.headerPtr[m]
					eptr = t.headerPtr[m+1]
					break
				}
				l = m + 1
			}
		}

	}

	if sptr == 0 {
		err = errors.New("not found")
		return
	}

	blockLen := eptr - sptr
	t.dbFileHandler.Seek(sptr, 0)
	index := make([]byte, blockLen+INDEX_BLOCK_LENGTH)
	t.dbFileHandler.Read(index)
	var dataptr int64
	h = blockLen / INDEX_BLOCK_LENGTH
	l = 0

	for l <= h {
		m := int64(l+h) >> 1
		p := m * INDEX_BLOCK_LENGTH
		sip := getLong(index, p)
		if ip < sip {
			h = m - 1
		} else {
			eip := getLong(index, p+4)
			if ip > eip {
				l = m + 1
			} else {
				dataptr = getLong(index, p+8)
				break
			}
		}
	}

	if dataptr == 0 {
		err = errors.New("not found")
		return
	}

	dataLen := (dataptr >> 24) & 0xFF
	dataPtr := dataptr & 0x00FFFFFF

	t.dbFileHandler.Seek(dataPtr, 0)
	data := make([]byte, dataLen)
	t.dbFileHandler.Read(data)
	ipInfo = getIpInfo(getLong(data, 0), data[4:])
	return
}

func getLong(b []byte, offset int64) int64 {

	val := int64(b[offset]) |
		int64(b[offset+1])<<8 |
		int64(b[offset+2])<<16 |
		int64(b[offset+3])<<24

	return val

}

func ip2long(IpStr string) (int64, error) {
	bits := strings.Split(IpStr, ".")
	if len(bits) != 4 {
		return 0, errors.New("ip format error")
	}

	var sum int64
	for i, n := range bits {
		bit, _ := strconv.ParseInt(n, 10, 64)
		sum += bit << uint(24-8*i)
	}

	return sum, nil
}
