package main

import (
	rinit "PZ_GameServer/pzTestFanCalc/init"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	// al "pzTestFanCalc/common/util/arrayList"
	xz "PZ_GameServer/server/game/room/xizhou"
)

func main() {

	defer func() {
		if err := recover(); err != nil {
			errStr := fmt.Sprintln("错误是:", err)
			saveFile([]byte(errStr), "result.txt")
		}
	}()

	huIndex := 0
	huCard := 0
	huCard = rinit.GetHuId()
	huIndex = rinit.GetSiteIndex()

	// huIndex = rinit.InitData.SiteIndex
	r := rinit.InitRoom(huIndex)
	// 结算控制器

	r.StlCtrl = xz.GetXiZhouStatement(
		r.Type,
		500,
		[]string{
		//r.SeatSeatss[0].UID,
		// r.[1].UID,
		// r.Seats[2].UID,
		// r.Seats[3].UID,
		},
		&r,
	)
	// resu := r.CheckHu(0, 1)
	// log.Printf("Msg : %d\n", resu)

	//结算
	totalResult := r.FanCalc(huIndex, huCard)

	result, _ := json.Marshal(totalResult.TotalMsg)

	log.Printf("Msg : %s\n", result)
	saveFile(result, "result.txt")

}

func saveFile(str []byte, fileName string) bool {

	err := ioutil.WriteFile(fileName, str, os.ModeAppend)
	if err != nil {
		log.Printf("file write failed err is  : %s\n", err)
		return false
	}
	return true
}
