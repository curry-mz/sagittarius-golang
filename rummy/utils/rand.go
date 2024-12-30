package utils

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

const (
	letters = "abcdefghijklmnopqrstuvwxyzAB CDEFGHIJKLMNOPQRSTUVWXYZ"
	// 6 bits to represent readmine letter index
	letterIdBits = 6
	// All 1-bits as many as letterIdBits
	letterIdMask = 1<<letterIdBits - 1
	letterIdMax  = 63 / letterIdBits
)

type randUtil struct {
}

func Rand() *randUtil {
	return &randUtil{}
}

func (t *randUtil) RandInterval(min, max int) int {
	if min == max {
		return min
	}

	if min < 0 {
		min = 0
	}

	if min > max {
		min, max = max, min
	}

	rand.Seed(time.Now().UnixNano())

	return rand.Intn(max-min) + min
}

func (t *randUtil) RandInterval64(min, max int64) int64 {
	if min == max {
		return min
	}

	if min < 0 {
		min = 0
	}

	if min > max {
		min, max = max, min
	}

	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(max-min) + min
}

// 随机附近5个B段IP
func (t *randUtil) RandIp(strIp string) (ip string) {
	ips := strings.Split(strIp, ".")
	if len(ips) != 4 {
		return
	}

	ip3, _ := strconv.Atoi(ips[2])
	i1 := t.RandInterval(ip3-2, ip3+2)
	if (i1 > 0) && (i1 < 255) {
		ip3 = i1
	}

	ip4 := t.RandInterval(1, 254)
	ip = fmt.Sprintf("%s.%s.%d.%d", ips[0], ips[1], ip3, ip4)
	return
}

// slice乱序
func (t *randUtil) RandSliceInt(src []int) (dst []int) {
	dst = make([]int, len(src))
	if len(src) == 0 {
		return
	}
	copy(dst, src)

	for i := len(dst) - 1; i > 0; i-- {
		num := rand.Intn(i + 1)
		dst[i], dst[num] = dst[num], dst[i]
	}
	return
}

// 返回随机字符串
func (t *randUtil) RandStr(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdMax letters!
	for i, cache, remain := n-1, src.Int63(), letterIdMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdMax
		}
		if idx := int(cache & letterIdMask); idx < len(letters) {
			b[i] = letters[idx]
			i--
		}
		cache >>= letterIdBits
		remain--
	}
	return *(*string)(unsafe.Pointer(&b))
}

func (t *randUtil) GetRandStr(n int) string {
	str := []rune("4ZSabLcefkiYRXmno7PQq5rstuMdTvw8xzA3B10CjDEFGplHIJK9N62yOVWhUg")
	lenStr := len(str)
	b := make([]rune, n)
	for i := range b {
		b[i] = str[rand.Intn(lenStr)]
	}
	return string(b)
}

func (t *randUtil) GenerateUniqueString() string {
	nano := time.Now().UnixNano()
	charset := `4ZSabLcefkiYRXmno7PQq5rstuMdTvw8xzA3B10CjDEFGplHIJK9N62yOVWhUg`
	length := int64(len(charset))
	var resBuf []byte
	for {
		i := nano % length
		resBuf = append(resBuf, charset[i])
		nano = nano / length
		if nano <= 0 {
			break
		}
	}
	return string(resBuf)
}

type Interval struct {
	Start int64
	Ended int64
}

// 动态生成数字区间
func (r *randUtil) generateIntervals(nums []int64) []Interval {
	if len(nums) == 0 {
		return nil
	}

	intervals := []Interval{}
	for i := 0; i < len(nums); i++ {
		start := int64(0)
		end := nums[i]
		if i > 0 {
			start = intervals[i-1].Ended
			end = start + nums[i]
		}

		intervals = append(intervals, Interval{Start: start + 1, Ended: end})
	}

	return intervals
}

func (r *randUtil) PointOfFall(totalNumber int64, probabilities []int64) int {
	intervals := r.generateIntervals(probabilities)
	fmt.Println("intervals", intervals)
	randNumber := r.RandInterval64(1, totalNumber)
	fmt.Println("PointOfFallRandNumber", randNumber)
	for i, interval := range intervals {
		if randNumber >= interval.Start && randNumber < interval.Ended {
			return i
		}
	}

	return -1
}

func (r randUtil) GenerateMonthFinish() string {
	arr := []string{}
	for i := 0; i < 31; i++ {
		arr = append(arr, "0")
	}
	monthStr := strings.Join(arr, "-") // explode
	return monthStr
}

// 动态生成数字区间
func (r *randUtil) generateIntervalsV2(nums map[int]int64) map[int]Interval {
	if len(nums) == 0 {
		return nil
	}

	intervals := map[int]Interval{}
	i := int64(0)
	for key, value := range nums {
		start := int64(i)
		end := i + int64(value)
		intervals[key] = Interval{Start: start + 1, Ended: end}
		i = i + int64(value)
	}
	return intervals
}

func (r *randUtil) PointOfFallV2(totalNumber int64, probabilities map[int]int64) int {
	intervals := r.generateIntervalsV2(probabilities)
	randNumber := rand.Int63n(totalNumber)
	for i, interval := range intervals {
		if randNumber >= interval.Start && randNumber < interval.Ended {
			return i
		}
	}

	return 0
}
func (r *randUtil) PlayGame(percentage int) (bool, error) {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())

	// 生成0到99的随机数，表示百分比的范围
	randomNumber := rand.Intn(100)

	// 判断是否胜利
	if randomNumber < percentage {
		return true, nil
	} else {
		return false, nil
	}
}
