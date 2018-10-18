package statement

// 面下杠就是玩家已经碰了三张一样的牌时，玩家自己又摸到了第四张一样的牌，这时候选择的杠牌  (碰后暗杠)
// 加杠就是指不止一种四张一样的，有好几种四张一样的
// 在胡牌后，可以继续选择面下杠或暗杠
// 抢杠：别人杠的牌自己能胡，可以抢杠，算做一般的点炮，基本胡型。

// ----------------------------

// 添加结算记录
func (s *StatementCtl) AddCalcRecord(score *OnceResult) {
	//s.RecordCalc.Add(score)
}

// 胡牌结算
func (s *StatementCtl) FanCalc(seatIndex int, arg ...interface{}) []int {

	// *****需要用到的变量*****
	// 方位
	// 当前风圈
	// 庄家
	// 手牌
	// 胡的牌
	// 自摸
	// 一炮多响
	// 听牌
	// 剩多少牌
	// 开始多少牌
	// 吃 碰 杠 暗杠 碰杠 胡  的数量和目标
	// 花牌数量
	// 杠牌的总数
	// (承包关系)

	return nil
}

// 结算一次分数(用于实时计算,  四川麻将, 杠牌后直接结算)
func (s *StatementCtl) ToolCalc() *OnceResult {
	//	score := []int{0, 0, 0, 0}
	//	msg := []string{"", "", "", ""}
	//	lastTool := s.GetPrvTool(0)

	//	switch lastTool.ToolType {

	//	case TEnum_AnKong: // 暗杠
	//		rz := 2 * s.BaseScore
	//		score[0] = -rz
	//		score[1] = -rz
	//		score[2] = -rz
	//		score[3] = -rz
	//		msg[0] = "被下雨"
	//		msg[1] = "被下雨"
	//		msg[2] = "被下雨"
	//		msg[3] = "被下雨"
	//		score[lastTool.Index] = rz * 3
	//		msg[lastTool.Index] = "下雨"

	//	case TEnum_Kong: // 直杠
	//		rz := 2 * s.BaseScore
	//		score[lastTool.Index] = rz
	//		score[lastTool.TIndex] = -rz
	//		msg[lastTool.Index] = "刮风"
	//		msg[lastTool.TIndex] = "被刮风"

	//	case TEnum_PengKong: // 碰杠
	//		rz := 1 * s.BaseScore
	//		score[0] = -rz
	//		score[1] = -rz
	//		score[2] = -rz
	//		score[3] = -rz
	//		msg[0] = "被下雨"
	//		msg[1] = "被下雨"
	//		msg[2] = "被下雨"
	//		msg[3] = "被下雨"
	//		score[lastTool.Index] = rz * 3
	//		msg[lastTool.Index] = "下雨"
	//	}

	//	oc := OnceScore{Score: score, MSG: msg}
	//	s.AddCalcRecord(&oc)
	return nil
}

//// 结束牌局后的结算
//func (s *StatementCtl) DrawCalc() *OnceScore {
//	score := []int{0, 0, 0, 0}
//	msg := []string{"", "", "", ""}

//	//	// 查大叫
//	//	tingScore := s.CalcDajiao(s.CheckTing())
//	//	fmt.Println("查大叫", tingScore)

//	//	// 查花猪
//	//	huazhuScore := s.CalcHuaZhu(s.CheckHuaZhu())
//	//	fmt.Println("查花猪", huazhuScore)

//	oc := OnceScore{Score: score, MSG: msg}
//	return &oc
//}

// 获得最后结算统计
func (s *StatementCtl) GetTotal() *OnceResult {

	//	score := []int{0, 0, 0, 0}
	//	msg := []string{"", "", "", ""}
	//	for i := 0; i < s.RecordCalc.Count; i++ {
	//		calc := (*s.RecordCalc.Index(i)).(*OnceScore)
	//		score[0] += calc.Score[0]
	//		score[1] += calc.Score[1]
	//		score[2] += calc.Score[2]
	//		score[3] += calc.Score[3]
	//		msg[0] += " " + calc.MSG[0]
	//		msg[1] += " " + calc.MSG[1]
	//		msg[2] += " " + calc.MSG[2]
	//		msg[3] += " " + calc.MSG[3]
	//	}
	//	oc := OnceScore{Score: score, MSG: msg}
	//	return &oc
	return nil
}

//// 计算花猪
//func (sc *StatementCtl) CalcHuaZhu(huazhu []int) []int {

//	//	score := []int{0, 0, 0, 0}
//	//	msg := []string{"", "", "", ""}
//	//	types := []int{0, 0, 0, 0}

//	//	l := len(huazhu)
//	//	s := sc.BaseScore * 64

//	//	for i := 0; i < l; i++ {
//	//		if huazhu[i] > 0 {
//	//			count := 0
//	//			for h := 0; h < l; h++ {
//	//				if huazhu[h] <= 0 && h != i {
//	//					score[h] += s
//	//					msg[h] = "被花猪"
//	//					//types[h] = 0
//	//					count++
//	//				}
//	//			}
//	//			score[i] -= s * count
//	//			msg[i] = "花猪"
//	//			fmt.Println("花猪", score[i])
//	//			types[i] = TEnum_HuaZhu // 花猪
//	//		}
//	//	}

//	//	oc := OnceScore{Score: score, MSG: msg, Type: types}
//	//	sc.AddCalcRecord(&oc)
//	//	return score
//	return nil
//}

//// 计算大叫
//func (sc *StatementCtl) CalcDajiao(dajiao []int) []int {
//	score := []int{0, 0, 0, 0}
//	//s := -sc.BaseScore * 64
//	for i := 0; i < len(dajiao); i++ {
//		if dajiao[i] <= 0 {
//			//			score[0] += s
//			//			score[1] += s
//			//			score[2] += s
//			//			score[3] += s
//			//			score[i] -= s * 3
//			//			sc.List[i].TotalScore += score[i]
//			//			oc := &OnceCalc{ReciveScore: score, MSG: "大叫"}
//			//			sc.List[i].EveryCalc.Add(oc)
//		}
//	}

//	return score
//}

//// 检查是否听牌(查大叫)
//func (sc *StatementCtl) CheckTing() []int {
//	tings := []int{0, 0, 0, 0}
//	for index := 0; index < len(sc.CCtl.Seats); index++ {
//		tings[index] = sc.room.Seat[index].Ting
//	}
//	return tings
//}
