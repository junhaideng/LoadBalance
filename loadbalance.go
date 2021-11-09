package main

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"sort"
	"time"
)

// 未考虑并发，仅提供一般算法思路

// ------------------------轮询法-----------------------
func RoundRobin(servers []string) func() string {
	pos := 0
	return func() string {
		res := servers[pos]
		pos++
		if pos == len(servers) {
			pos = 0
		}
		return res
	}
}

//--------------------随机法---------------------------
func init() {
	// 随机数种子初始化
	rand.Seed(time.Now().UnixNano())
}

func Random(servers []string) string {
	index := rand.Intn(len(servers))
	return servers[index]
}

//--------------------权重轮询--------------------------

type Server struct {
	ip     string
	weight int
}

// 方法一
func WeigthRoundRobin(servers []Server) func() string {
	pos := 0
	s := make([]string, 0, len(servers))

	for i := 0; i < len(servers); i++ {
		for j := 0; j < servers[i].weight; j++ {
			s = append(s, servers[i].ip)
		}
	}
	return func() string {
		res := s[pos]
		pos++
		if pos == len(s) {
			pos = 0
		}
		return res
	}
}

// 方法二
func SmoothWeightRoundRobin(servers []Server) func() string {
	current := make([]int, len(servers))
	total := 0
	for i := 0; i < len(servers); i++ {
		total += servers[i].weight
	}

	return func() string {
		max := 0
		index := 0
		// 加入 effective 权重
		// 并且找到 current 权重最大的
		for i := 0; i < len(servers); i++ {
			current[i] += servers[i].weight
			if current[i] > max {
				max = current[i]
				index = i
			}
		}

		current[index] -= total
		return servers[index].ip
	}
}

func testSmoothWeightRoundRobin() {
	s := []Server{
		{ip: "192.168.0.1", weight: 3},
		{ip: "192.168.0.2", weight: 1},
	}

	get := SmoothWeightRoundRobin(s)
	memo := make(map[string]int)
	for i := 0; i < 10000; i++ {
		memo[get()]++
	}
	fmt.Println(memo)

}

//-----------------------权重随机---------------------
// 方法一
func WeigthRandom(servers []Server) func() string {
	rand.Seed(time.Now().UnixNano())

	s := make([]string, 0, len(servers))

	for i := 0; i < len(servers); i++ {
		for j := 0; j < servers[i].weight; j++ {
			s = append(s, servers[i].ip)
		}
	}

	return func() string {
		index := rand.Intn(len(s))
		return s[index]
	}
}

// 方法二
func WeigthRandom2(servers []Server) func() string {
	rand.Seed(time.Now().UnixNano())

	preSum := make([]int, len(servers)+1)
	for i := 1; i < len(servers)+1; i++ {
		preSum[i] = preSum[i-1] + servers[i-1].weight
	}

	total := preSum[len(preSum)-1]

	return func() string {
		num := rand.Intn(total) + 1
		index := sort.Search(len(preSum), func(i int) bool {
			return preSum[i] >= num
		})

		return servers[index-1].ip
	}
}

func testWeightRandom() {
	s := []Server{
		{ip: "192.168.0.1", weight: 3},
		{ip: "192.168.0.2", weight: 2},
	}

	get := WeigthRandom2(s)
	memo := make(map[string]int)
	for i := 0; i < 100000; i++ {
		memo[get()]++
	}
	fmt.Println(memo)
}

// ---------------------哈希法----------------------
func HashLoadBalance(servers []string) func(key string) string {
	return func(key string) string {
		f := fnv.New32()
		f.Write([]byte(key))
		// 计算 hash 值
		h := f.Sum32()
		// 取模获取对应的服务器
		return servers[int(h)%len(servers)]
	}
}

// -------------------最小连接数-------------------

// -------------------最小响应时间-----------------
// 响应时间统计
type RT struct {
	count int           // 多少次
	total time.Duration // 总时间
}

type LoadBalancer struct {
	servers []string
	rt      map[string]RT // 响应时间
}

func (l *LoadBalancer) LeastResponseTime() string {
	if len(l.servers) == 0 {
		return ""
	}
	// 初始化
	server := l.servers[0]
	min := l.rt[server].total / time.Duration(l.rt[server].count)

	// 选择最小的
	for i := 1; i < len(l.servers); i++ {
		rt := l.rt[l.servers[i]]
		average := rt.total / time.Duration(rt.count)
		if average < min {
			min = average
			server = l.servers[i]
		}
	}

	return server
}

func testLeastResponseTime() {
	l := LoadBalancer{
		servers: []string{"192.168.0.1", "192.168.0.2"},
		rt: map[string]RT{
			"192.168.0.1": {
				10, 1000,
			},
			"192.168.0.2": {
				10, 2000,
			},
		},
	}
	fmt.Println(l.LeastResponseTime())
}

func main() {
	// testWeightRandom()
	testLeastResponseTime()
}
