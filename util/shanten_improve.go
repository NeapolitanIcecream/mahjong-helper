package util

import (
	"fmt"
	"sort"
)

// map[改良牌]进张
type Improves map[int]Waits

// 1/4/7/10/13 张手牌的分析结果
type WaitsWithImproves13 struct {
	// 手牌
	Tiles34 []int

	// 向听数
	Shanten int

	// 进张：摸到这张牌可以让向听数前进
	Waits Waits

	// map[进张牌]向听前进后的进张数（这里让向听前进的切牌是最优切牌，即让向听前进后的进张数最大的切牌）
	NextShantenWaitsCountMap map[int]int

	// 改良：摸到这张牌虽不能让向听数前进，但可以让进张变多
	Improves Improves

	// 改良情况数
	ImproveWayCount int

	// 对于每张牌，摸到之后的手牌进张数（如果摸到的是 Waits 中的牌，则进张数视作摸到之前的进张数）
	ImproveWaitsCount34 []int

	// 在没有摸到进张时的改良进张数的加权均值
	AvgImproveWaitsCount float64

	// 向听前进后的进张数的加权均值
	AvgNextShantenWaitsCount float64
}

// avgImproveWaitsCount: 在没有摸到进张时的改良进张数的加权均值
func (r *WaitsWithImproves13) analysis() (avgImproveWaitsCount float64, avgNextShantenWaitsCount float64) {
	const leftTile = 4

	if len(r.Improves) > 0 {
		improveScore := 0
		weight := 0
		for i := 0; i < 34; i++ {
			w := leftTile - r.Tiles34[i]
			improveScore += w * r.ImproveWaitsCount34[i]
			weight += w
		}
		avgImproveWaitsCount = float64(improveScore) / float64(weight)
		r.AvgImproveWaitsCount = avgImproveWaitsCount
	} else {
		r.AvgImproveWaitsCount = float64(r.Waits.AllCount())
	}

	if len(r.NextShantenWaitsCountMap) > 0 {
		nextShantenWaitsSum := 0
		weight := 0
		for tile, c := range r.NextShantenWaitsCountMap {
			w := leftTile - r.Tiles34[tile]
			nextShantenWaitsSum += w * c
			weight += w
		}
		avgNextShantenWaitsCount = float64(nextShantenWaitsSum) / float64(weight)
		r.AvgNextShantenWaitsCount = avgNextShantenWaitsCount
	}

	return
}

// 调试用
func (r *WaitsWithImproves13) String() string {
	s := fmt.Sprintf("%s\n%.2f [%d 改良]",
		r.Waits.String(),
		r.AvgImproveWaitsCount,
		r.ImproveWayCount,
	)
	if r.Shanten > 0 {
		s += fmt.Sprintf(" %.2f %s进张",
			r.AvgNextShantenWaitsCount,
			NumberToChineseShanten(r.Shanten-1),
		)
	}
	return s
}

// 1/4/7/10/13 张牌，计算向听数和进张
func CalculateShantenAndWaits13(tiles34 []int, isOpen bool) (shanten int, waits Waits) {
	shanten = CalculateShanten(tiles34, isOpen)

	const leftTile = 4

	// 剪枝：检测非浮牌，在不考虑国士无双的情况下，这种牌是不可能让向听数前进的（但有改良的可能，不过 CalculateShantenAndWaits13 函数不考虑这个）
	// 此处优化提升了约 30% 的性能
	//needCheck34 := make([]bool, 34)
	//idx := -1
	//for i := 0; i < 3; i++ {
	//	for j := 0; j < 9; j++ {
	//		idx++
	//		if tiles34[idx] == 0 {
	//			continue
	//		}
	//		if j == 0 {
	//			needCheck34[idx] = true
	//			needCheck34[idx+1] = true
	//			needCheck34[idx+2] = true
	//		} else if j == 1 {
	//			needCheck34[idx-1] = true
	//			needCheck34[idx] = true
	//			needCheck34[idx+1] = true
	//			needCheck34[idx+2] = true
	//		} else if j < 7 {
	//			needCheck34[idx-2] = true
	//			needCheck34[idx-1] = true
	//			needCheck34[idx] = true
	//			needCheck34[idx+1] = true
	//			needCheck34[idx+2] = true
	//		} else if j == 7 {
	//			needCheck34[idx-2] = true
	//			needCheck34[idx-1] = true
	//			needCheck34[idx] = true
	//			needCheck34[idx+1] = true
	//		} else {
	//			needCheck34[idx-2] = true
	//			needCheck34[idx-1] = true
	//			needCheck34[idx] = true
	//		}
	//	}
	//}
	//for i := 27; i < 34; i++ {
	//	if tiles34[i] > 0 {
	//		needCheck34[i] = true
	//	}
	//}

	waits = Waits{}
	for i := 0; i < 34; i++ {
		//if !needCheck34[i] {
		//	continue
		//}
		// 摸牌
		tiles34[i]++
		if newShanten := CalculateShanten(tiles34, isOpen); newShanten < shanten {
			// 向听前进了，则换的这张牌为进张
			waits[i] = leftTile - (tiles34[i] - 1)
		}
		tiles34[i]--
	}
	return
}

// 1/4/7/10/13 张牌，计算向听数、进张、改良等
func CalculateShantenWithImproves13(tiles34 []int, isOpen bool) (waitsWithImproves *WaitsWithImproves13) {
	shanten13, waits := CalculateShantenAndWaits13(tiles34, isOpen)
	waitsCount := waits.AllCount()

	//fmt.Println(Tiles34ToMergedStrWithBracket(tiles34), waits)

	nextShantenWaitsCountMap := map[int]int{} // map[进张牌]听多少张牌
	improves := Improves{}
	improveWayCount := 0
	improveWaitsCount34 := make([]int, 34)
	// 初始化成基本进张
	for i := 0; i < 34; i++ {
		improveWaitsCount34[i] = waitsCount
	}

	const leftTile = 4

	for i := 0; i < 34; i++ {
		if tiles34[i] == leftTile {
			continue
		}
		// 摸牌
		tiles34[i]++
		if _, ok := waits[i]; ok {
			// 是进张
			for j := 0; j < 34; j++ {
				if tiles34[j] == 0 || j == i {
					continue
				}
				// 切牌
				tiles34[j]--
				// 正确的切牌
				if newShanten13, _waits := CalculateShantenAndWaits13(tiles34, isOpen); newShanten13 < shanten13 {
					// 切牌一般切进张最多的
					if waitsCount := _waits.AllCount(); waitsCount > nextShantenWaitsCountMap[i] {
						nextShantenWaitsCountMap[i] = waitsCount
					}
				}
				tiles34[j]++
			}
		} else {
			// 不是进张，但可能有改良
			for j := 0; j < 34; j++ {
				if tiles34[j] == 0 || j == i {
					continue
				}
				// 切牌
				tiles34[j]--
				// 正确的切牌
				if newShanten13, improveWaits := CalculateShantenAndWaits13(tiles34, isOpen); newShanten13 == shanten13 {
					// 若进张数变多，则为改良
					if improveWaitsCount := improveWaits.AllCount(); improveWaitsCount > waitsCount {
						improveWayCount++
						if improveWaitsCount > improveWaitsCount34[i] {
							improveWaitsCount34[i] = improveWaitsCount
							improves[i] = improveWaits
						}
						//fmt.Println(fmt.Sprintf("    摸 %s 切 %s 改良:", mahjongZH[drawTile], mahjongZH[discardTile]), improveWaitsCount, TilesToMergedStrWithBracket(improveWaits.indexes()))
					}
				}
				tiles34[j]++
			}
		}
		tiles34[i]--
	}

	_tiles34 := make([]int, 34)
	copy(_tiles34, tiles34)
	waitsWithImproves = &WaitsWithImproves13{
		Tiles34: _tiles34,
		Shanten: shanten13,
		Waits:   waits,
		NextShantenWaitsCountMap: nextShantenWaitsCountMap,
		Improves:                 improves,
		ImproveWayCount:          improveWayCount,
		ImproveWaitsCount34:      improveWaitsCount34,
	}
	waitsWithImproves.analysis()

	return
}

type WaitsWithImproves14 struct {
	Result13 *WaitsWithImproves13
	// 需要切的牌
	DiscardTile int
	// 切掉这张牌后的向听数
	Shanten int
}

func (r *WaitsWithImproves14) String() string {
	return fmt.Sprintf("切 %s: %s", mahjongZH[r.DiscardTile], r.Result13.String())
}

type WaitsWithImproves14List []*WaitsWithImproves14

func (l WaitsWithImproves14List) Sort() {
	sort.Slice(l, func(i, j int) bool {
		ri, rj := l[i].Result13, l[j].Result13

		allCountRate := float64(ri.Waits.AllCount()) / float64(rj.Waits.AllCount())
		if allCountRate < 1 {
			allCountRate = 1 / allCountRate
		}
		if allCountRate > 1.1 {
			return ri.Waits.AllCount() > rj.Waits.AllCount()
		}

		// 相差在 10% 以内的，向听前进后的进张数的加权均值多者为优
		if ri.AvgNextShantenWaitsCount != rj.AvgNextShantenWaitsCount {
			return ri.AvgNextShantenWaitsCount > rj.AvgNextShantenWaitsCount
		}

		if ri.AvgImproveWaitsCount != rj.AvgImproveWaitsCount {
			return ri.AvgImproveWaitsCount > rj.AvgImproveWaitsCount
		}

		if ri.ImproveWayCount != rj.ImproveWayCount {
			return ri.ImproveWayCount > rj.ImproveWayCount
		}

		return l[i].DiscardTile > l[j].DiscardTile
	})
}

// 2/5/8/11/14 张牌，计算向听数、进张、改良、向听倒退等
func CalculateShantenWithImproves14(tiles34 []int, isOpen bool) (shanten int, waitsWithImproves WaitsWithImproves14List, incShantenResults WaitsWithImproves14List) {
	shanten = CalculateShanten(tiles34, isOpen)

	for i := 0; i < 34; i++ {
		if tiles34[i] == 0 {
			continue
		}
		tiles34[i]-- // 切牌
		result13 := CalculateShantenWithImproves13(tiles34, isOpen)
		r := &WaitsWithImproves14{
			Result13:    result13,
			DiscardTile: i,
			Shanten:     result13.Shanten,
		}
		if result13.Shanten == shanten {
			waitsWithImproves = append(waitsWithImproves, r)
		} else {
			// 向听倒退
			incShantenResults = append(incShantenResults, r)
		}
		tiles34[i]++
	}
	waitsWithImproves.Sort()
	incShantenResults.Sort()
	return
}
