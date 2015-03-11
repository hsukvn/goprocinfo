package linux

import (
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
)

type SockStat struct {
	// sockets:
	SocketsUsed uint64 `json:"sockets_used" altname:"sockets:used"`

	// TCP:
	TCPInUse     uint64 `json:"tcp_in_use" altname:"TCP:inuse"`
	TCPOrphan    uint64 `json:"tcp_orphan" altname:"TCP:orphan"`
	TCPTimeWait  uint64 `json:"tcp_time_wait" altname:"TCP:tw"`
	TCPAllocated uint64 `json:"tcp_allocated" altname:"TCP:alloc"`
	TCPMemory    uint64 `json:"tcp_memory" altname:"TCP:mem"`

	// UDP:
	UDPInUse  uint64 `json:"udp_in_use" altname:"UDP:inuse"`
	UDPMemory uint64 `json:"udp_memory" altname:"UDP:mem"`

	// UDPLITE:
	UDPLITEInUse uint64 `json:"udplite_in_use" altname:"UDPLITE:inuse"`

	// RAW:
	RAWInUse uint64 `json:"raw_in_use" altname:"RAW:inuse"`

	// FRAG:
	FRAGInUse  uint64 `json:"frag_in_use" altname:"FRAG:inuse"`
	FRAGMemory uint64 `json:"frag_memory" altname:"FRAG:memory"`
}

func ReadSockStat(path string) (*SockStat, error) {
	data, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")

	// Maps a meminfo metric to its value (i.e. MemTotal --> 100000)
	statMap := make(map[string]uint64)

	var sockStat SockStat = SockStat{}

	for _, line := range lines {
		if strings.Index(line, ":") == -1 {
			continue
		}

		statType := line[0 : strings.Index(line, ":")+1]

		// The fields have this pattern: inuse 27 orphan 1 tw 23 alloc 31 mem 3
		// The stats are grouped into pairs and need to be parsed and placed into the stat map.
		key := ""
		for k, v := range strings.Fields(line[strings.Index(line, ":")+1:]) {
			// Every second field is a value.
			if (k+1)%2 != 0 {
				key = v
				continue
			}
			val, _ := strconv.ParseUint(v, 10, 64)
			statMap[statType+key] = val
		}
	}

	elem := reflect.ValueOf(&sockStat).Elem()
	typeOfElem := elem.Type()

	for i := 0; i < elem.NumField(); i++ {
		val, ok := statMap[typeOfElem.Field(i).Tag.Get("altname")]
		if ok {
			elem.Field(i).SetUint(val)
		}
	}

	return &sockStat, nil
}
