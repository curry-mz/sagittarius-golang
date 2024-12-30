package game

import (
	"code.cd.local/games-go/mini/rummy.busi/model"
	"math/rand"
	"sort"
	"time"
)

var _ran = rand.New(rand.NewSource(time.Now().UnixNano()))

func GetCards(winType int) (int, []uint8, int, []uint8) {
	// 赢的牌
	win := getCards(winType)
	var winValue int
	if winType == model.HighCard {
		winValue = getHighCardValue(win)
	}

	// 输的牌
	var lose []uint8
	var loseValue int
	var loseType = model.HighCard
	for {
		lose = getCards(loseType)
		loseValue = getHighCardValue(lose)
		if winType == model.HighCard && winValue == loseValue {
			continue
		}

		// 是否有一样的牌
		same := false
		for _, wc := range win {
			for _, lc := range lose {
				if wc == lc {
					same = true
					break
				}
			}
		}
		if !same {
			break
		}
	}
	if winType == model.HighCard && winValue < loseValue {
		win, lose = lose, win
	}
	return winType, win, loseType, lose
}

func getCards(t int) []uint8 {
	switch t {
	case model.HighCard:
		return getHighCard()
	case model.Pair:
		return getPair()
	case model.BigPair:
		return getBigPair()
	case model.Color:
		return getColor()
	case model.Run:
		return getRun()
	case model.PureRun:
		return getPureRun()
	case model.Trail:
		return getTrail()
	}
	return nil
}

func getTrail() []uint8 {
	var cards []uint8
	v := _ran.Intn(13) + 1
	colors := []int{0, 1, 2, 3}
	_ran.Shuffle(len(colors), func(i, j int) {
		colors[i], colors[j] = colors[j], colors[i]
	})
	for i := 0; i < 3; i++ {
		cards = append(cards, uint8(colors[i]<<4)|uint8(v))
	}
	return cards
}

func getPureRun() []uint8 {
	var cards []uint8
	v := _ran.Intn(12) + 1
	colors := []int{0, 1, 2, 3}
	_ran.Shuffle(len(colors), func(i, j int) {
		colors[i], colors[j] = colors[j], colors[i]
	})
	for i := 0; i < 3; i++ {
		value := v + i
		if value > 13 {
			value = 1
		}
		cards = append(cards, uint8(colors[0]<<4)|uint8(value))
	}
	_ran.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})
	return cards
}

func getRun() []uint8 {
	var cards []uint8
	v := _ran.Intn(12) + 1
	colors := []int{0, 1, 2, 3}
	_ran.Shuffle(len(colors), func(i, j int) {
		colors[i], colors[j] = colors[j], colors[i]
	})
	for i := 0; i < 3; i++ {
		value := v + i
		if value > 13 {
			value = 1
		}
		cards = append(cards, uint8(colors[i]<<4)|uint8(value))
	}
	_ran.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})
	return cards
}

func getColor() []uint8 {
	var cards []uint8
	color := _ran.Intn(4)
	values := [][]int{
		{1, 2, 4, 5, 7, 8, 10, 11, 13},
		{2, 3, 5, 6, 8, 9, 11, 12},
		{1, 3, 4, 6, 7, 9, 10, 13},
	}
	value := values[_ran.Intn(len(values))]
	_ran.Shuffle(len(value), func(i, j int) {
		value[i], value[j] = value[j], value[i]
	})
	for i := 0; i < 3; i++ {
		v := value[i]
		cards = append(cards, uint8(color<<4)|uint8(v))
	}
	return cards
}

func getBigPair() []uint8 {
	var cards []uint8
	value := _ran.Intn(6) + 9
	if value == 14 {
		value = 1
	}
	colors := []int{0, 1, 2, 3}
	_ran.Shuffle(len(colors), func(i, j int) {
		colors[i], colors[j] = colors[j], colors[i]
	})
	for i := 0; i < 2; i++ {
		cards = append(cards, uint8(colors[i]<<4)|uint8(value))
	}
	sValue := (_ran.Intn(11) + 1) + value
	if sValue > 13 {
		sValue -= 13
	}
	cards = append(cards, uint8(colors[_ran.Intn(len(colors))]<<4)|uint8(sValue))
	return cards
}

func getPair() []uint8 {
	var cards []uint8
	value := _ran.Intn(7) + 2
	colors := []int{0, 1, 2, 3}
	_ran.Shuffle(len(colors), func(i, j int) {
		colors[i], colors[j] = colors[j], colors[i]
	})
	for i := 0; i < 2; i++ {
		cards = append(cards, uint8(colors[i]<<4)|uint8(value))
	}
	sValue := (_ran.Intn(11) + 1) + value
	if sValue > 13 {
		sValue -= 13
	}
	cards = append(cards, uint8(colors[_ran.Intn(len(colors))]<<4)|uint8(sValue))
	return cards
}

func getHighCard() []uint8 {
	var cards []uint8

	values := [][]int{
		{1, 2, 4, 5, 7, 8, 10, 11, 13},
		{2, 3, 5, 6, 8, 9, 11, 12},
		{1, 3, 4, 6, 7, 9, 10, 13},
	}
	value := values[_ran.Intn(len(values))]
	_ran.Shuffle(len(value), func(i, j int) {
		value[i], value[j] = value[j], value[i]
	})

	colors := []int{0, 1, 2, 3}
	_ran.Shuffle(len(colors), func(i, j int) {
		colors[i], colors[j] = colors[j], colors[i]
	})

	for i := 0; i < 2; i++ {
		cards = append(cards, uint8(colors[i]<<4)|uint8(value[i]))
	}
	cards = append(cards, uint8(colors[_ran.Intn(len(colors))]<<4)|uint8(value[2]))
	return cards
}

func getHighCardValue(cards []uint8) int {
	tmp := []int{int(cards[0] & 0xF), int(cards[1] & 0xF), int(cards[2] & 0xF)}
	if tmp[0] == 1 {
		tmp[0] = 14
	} else if tmp[1] == 1 {
		tmp[1] = 14
	} else if tmp[2] == 1 {
		tmp[2] = 14
	}
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i] > tmp[j]
	})
	return tmp[0]*10000 + tmp[1]*100 + tmp[2]
}
