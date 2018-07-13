package main

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"io/ioutil"
	"bytes"
	"strconv"
	"os"
)

var TOKEN = "TOKEN"

var cookieJar, _ = cookiejar.New(nil)
var client = &http.Client{
Jar: cookieJar,
}
var postURL = "https://apiv2.twitcasting.tv/internships/2018/games/"

type ID struct{
	Id string `json:"id"`
}

type Result struct{
	Hit int `json:"hit"`
	Message string `json:"message"`
}

//残り未確定位置
var  left = []int{0,1,2,3,4,5,6,7,8,9}
//確定保存用
var con = []int{0,1,2,3,4,5,6,7,8,9}

func main() {
	getURL := "https://apiv2.twitcasting.tv/internships/2018/games?level=10"
	req, _ := http.NewRequest("GET", getURL, nil)
	req.Header.Set("Authorization", "Bearer "+TOKEN)
	res, _ := client.Do(req)

	defer res.Body.Close()

	var idJson ID
	byteArray, _ := ioutil.ReadAll(res.Body)
	json.Unmarshal(byteArray, &idJson)
	//debug
	// println(string(byteArray))
	postURL += idJson.Id

	//debug
	// println(postURL)

	//入れ替え時のflag用
	notHit := false
	//探索時にHitが減った場合の保存用
	minus := []int{}
	//探索時にHitが増えた場合の保存用
	plus := []int{}

	//初期Hit数
	defHit := answerString("0123456789")

	//入れ替え後のHit数保存用
	conHit := defHit
	for next := 1; ; next ++{
		/*
		新しいanswer配列の準備
		残り未確定を順番に入れ替えて探索
		*/
		ans := swap(con, left[0], left[next])

		//新しいHit数とそれ以前のHit数を比較しSwitch
			switch answerList(ans) - conHit {
			case -2:
				left = removePosition(left, 0, next)
				notHit = false
				next = 0

				minus = []int{}
				plus = []int{}
				break

			case -1:
				if notHit {
					left = removePosition(left, next)

					next --
					break
				}
				minus = append(minus, left[next])
				if (len(minus) > defHit) {
					left = removePosition(left, 0)
					notHit = false
					minus = []int{}

					next = 0
				}
				break
			case 0:
				if len(minus) != 0 {
					for _, v := range minus {
						left = removeObject(left, v)
					}
					minus = []int{}
					next --
				}
				notHit = true
				break
			case 1:
				if len(minus) != 0 {
					for _, v := range minus {
						left = removeObject(left, v)
					}
					minus = []int{}
					next --
				}

				plus = append(plus, left[next])
				if len(plus) == 2 {
					add := conSearch(plus, conHit)
					conHit += add

					plus = []int{}
					notHit = false
					next = 0
				}
				notHit = true
				break

			case 2:
				con = swap(con, left[0], left[next])

				left = removePosition(left, 0, next)
				notHit = false
				conHit += 2
				next = 0
				minus = []int{}
				plus = []int{}
				break

			}


			//debug
			for _, v := range left {
				print(v)
			}
			println()


			if next == len(left) - 1 {
				for _, v := range minus {
					left = removeObject(left, v)
				}

				minus = []int{}
				next = 0
				notHit = false
			}
		}

	/* 先頭からの入れ替え探索と後からの探索で並列化すれば早くなるかと思ったがやめた
		wg := sync.WaitGroup{}
		wg.Add(2)
		println("2段階")
		//先頭から入れ替え
		go func() {
			defer wg.Done()
			for next := 1; next < len(left);{
				ans := swap(con, left[0], left[next])
        
			//debug
			//println("conHit=",conHit)

				switch answerList(ans) - conHit {
				case 2:
					con = swap(con, left[0], left[next])

					left = removePosition(left, 0, next)
					next = 0
					plus = []int{}
					conHit += 2
					break

				case 1:
					plus = append(plus, left[next])
					if (len(plus) == 2) {
						conHit += conSearch(plus, conHit)

						plus = []int{}
						next = 0
					}
					break
				case 0:
					break
				}

				next ++
			}
		}()
	*/

	//debug
	for _, v := range left {
		print(v)
	}


}


/*
2度 +1Hitとなったときの探索関数 返り値は確定した数
例:
answer  ->	hit
012 	-> 	0	初期値
102		->	1	0&1入れ替え
210		-> 	1	0&2入れ替え
ここでconsearch呼び出し
201		-> 	3	->	全て確定させreturn
201		->	2	->	0番目と1番目を確定させreturn
201		->	0	->	'120'とし,0番目と2番目を確定させreturn
 */
func conSearch(ints []int, defaultHit int) int{
	ans := swap(swap(con, left[0], ints[0]), left[0], ints[1])
	switch answerList(ans) - defaultHit{
	case 3:
		con = swap(swap(con, left[0], ints[0]), left[0], ints[1])
		left = removeObject(left, left[0], ints[0], ints[1])

		return 3
	case 2:
		con = swap(swap(con, left[0], ints[0]), left[0], ints[1])
		left = removeObject(left, left[0], ints[0])

		return 2
	case 0:
		con = swap(swap(con, left[0], ints[1]), left[0], ints[0])
		left = removeObject(left, left[0], ints[1])

		return 2
	}

	println("error:-1 can't search")
	os.Exit(-1)
	return -1
}

/*
answerがStringでのPOST
下記の配列よりは断然早いと思うが仕様上,初期値以外配列の方しか使っていない
 */
func answerString(answer string) (int){
	var data = []byte(`{"answer":"`+answer+`"}`)
//	debug
 println(string(data))
	req, _ := http.NewRequest("POST", postURL, bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+ TOKEN)
	res, _ := client.Do(req)

	defer res.Body.Close()

	var result Result
	byteArray, _ := ioutil.ReadAll(res.Body)
	json.Unmarshal(byteArray, &result)
	println(string(byteArray))

	return result.Hit
}

/*
answerがint配列でのPOST
処理は上記より遅いが仕様上致し方ない
 */
func answerList(list []int) (int){
	//少しでも速度上昇のため一度の代入文で結合
	var data = []byte(`{"answer":"`+strconv.Itoa(list[0])+strconv.Itoa(list[1])+strconv.Itoa(list[2])+strconv.Itoa(list[3])+strconv.Itoa(list[4])+strconv.Itoa(list[5])+strconv.Itoa(list[6])+strconv.Itoa(list[7])+strconv.Itoa(list[8])+strconv.Itoa(list[9])+`"}`)
//	debug
 	println(string(data))
	req, _ := http.NewRequest("POST", postURL, bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+ TOKEN)
	res, _ := client.Do(req)

	defer res.Body.Close()

	var result Result
	byteArray, _ := ioutil.ReadAll(res.Body)
	json.Unmarshal(byteArray, &result)
	println(string(byteArray))

	return result.Hit
}

/*
配列から指定した要素の削除
 */
func removeObject(ints []int, search ...int) []int {
	result := []int{}

	label:
	for _, v := range ints {
		for _, s := range search {
			if v == s {
				continue label
			}
		}
		result = append(result, v)
	}
	return result
}

/*
配列から指定した位置の要素を削除
 */
func removePosition(ints []int, pos ...int) []int {
	result := []int{}

	label:
	for i, v := range ints{
		for _, v := range pos{
			if i == v{
				continue label
			}
		}
		result = append(result, v)
	}
	return result
}


/*
配列の指定した要素を入れ替え
 */
func swap(ints []int, pos1 int, pos2 int) []int{
	intscpy := make([]int, len(ints))
	copy(intscpy, ints)
	intscpy[pos1], intscpy[pos2] = intscpy[pos2], intscpy[pos1]

	return intscpy
}
