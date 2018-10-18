package sdk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

//文件作用:微信请求接口
var (
	url           = "https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN"
	checkTokenUrl = "https://api.weixin.qq.com/sns/auth?access_token=%s&openid=%s"
)

type UserInfo struct {
	Openid     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	Headimgurl string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	Unionid    string   `json:"unionid"`
}

type Response struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

//拉取用户信息
func GetUserInfo(openid, token string) (*UserInfo, error) {
	var userInfo UserInfo

	requestUrl := fmt.Sprintf(url, token, openid)

	response, err := http.Get(requestUrl)
	defer response.Body.Close()

	if err != nil {
		return nil, err
	}

	data, _ := ioutil.ReadAll(response.Body)

	err = json.Unmarshal(data, &userInfo)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}

//校验token
func CheckTokenIsValid(openid, token string) (*Response, error) {
	var r Response
	requestUrl := fmt.Sprintf(checkTokenUrl, token, openid)
	response, err := http.Get(requestUrl)
	defer response.Body.Close()

	if err != nil {
		return nil, err
	}

	data, _ := ioutil.ReadAll(response.Body)

	err = json.Unmarshal(data, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}
