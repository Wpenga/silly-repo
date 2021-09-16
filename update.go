package jd_cookie

import (
	"fmt"
	"strings"

	"github.com/cdle/sillyGirl/core"
	"github.com/cdle/sillyGirl/develop/qinglong"
	"github.com/cdle/sillyGirl/im"
)

var pinQQ = core.NewBucket("pinQQ")
var pinTG = core.NewBucket("pinTG")

func init() {
	core.AddCommand("", []core.Function{
		{
			Rules:   []string{`raw pt_key=([^;=\s]+);\s*pt_pin=([^;=\s]+)`},
			FindAll: true,
			Handle: func(s im.Sender) interface{} {
				ck := &JdCookie{
					PtKey: s.Get(0),
					PtPin: s.Get(1),
				}
				if !ck.Available() {
					return "无效的ck，请重试。"
				}
				value := fmt.Sprintf("pt_key=%s;pt_pin=%s;", ck.PtKey, ck.PtPin)
				envs, err := qinglong.GetEnvs(fmt.Sprintf("pt_pin=%s;", ck.PtPin))
				if err != nil {
					return err
				}
				if s.GetImType() == "qq" {
					pinQQ.Set(ck.PtPin, s.GetUserID())
				}
				if s.GetImType() == "tg" {
					pinTG.Set(ck.PtPin, s.GetUserID())
				}
				if len(envs) == 0 {
					if err := qinglong.AddEnv(qinglong.Env{
						Name:  "JD_COOKIE",
						Value: value,
					}); err != nil {
						return err
					}
					return ck.Nickname + ",添加成功。"
				} else {
					env := envs[0]
					env.Value = value
					if err := qinglong.UdpEnv(env); err != nil {
						return err
					}
					return ck.Nickname + ",更新成功。"
				}
			},
		},
		{
			Rules:   []string{`raw pin=([^;=\s]+);\s*wskey=([^;=\s]+)`},
			FindAll: true,
			Handle: func(s im.Sender) interface{} {
				value := fmt.Sprintf("pin=%s;wskey=%s;", s.Get(0), s.Get(1))
				pt_key, err := getKey(value)
				if err == nil {
					if strings.Contains(pt_key, "fake") {
						return "无效的wskey，请重试。"
					}
				} else {
					s.Reply(err)
				}
				ck := &JdCookie{
					PtKey: pt_key,
					PtPin: s.Get(0),
				}
				ck.Available()
				envs, err := qinglong.GetEnvs(fmt.Sprintf("pin=%s;", ck.PtPin))
				if err != nil {
					return err
				}
				if s.GetImType() == "qq" {
					pinQQ.Set(s.Get(1), s.GetUserID())
				}
				if s.GetImType() == "tg" {
					pinTG.Set(s.Get(1), s.GetUserID())
				}
				var envCK *qinglong.Env
				var envWsCK *qinglong.Env
				for i := range envs {
					if strings.Contains(envs[i].Value, fmt.Sprintf("pin=%s;wskey=", ck.PtPin)) && envs[i].Name == "JD_WSCK" {
						envWsCK = &envs[i]
					} else if strings.Contains(envs[i].Value, fmt.Sprintf("pt_pin=%s;", ck.PtPin)) && envs[i].Name == "JD_COOKIE" {
						envCK = &envs[i]
					}
				}
				value2 := fmt.Sprintf("pt_key=%s;pt_pin=%s;", ck.PtKey, ck.PtPin)
				if envCK == nil {
					qinglong.AddEnv(qinglong.Env{
						Name:  "JD_COOKIE",
						Value: value2,
					})
				} else {
					envCK.Value = value2
					if err := qinglong.UdpEnv(*envCK); err != nil {
						return err
					}
				}
				if envWsCK == nil {
					if err := qinglong.AddEnv(qinglong.Env{
						Name:  "JD_WSCK",
						Value: value,
					}); err != nil {
						return err
					}
					return ck.Nickname + ",添加成功。"
				} else {
					env := envs[0]
					env.Value = value
					if err := qinglong.UdpEnv(env); err != nil {
						return err
					}
					return ck.Nickname + ",更新成功。"
				}
			},
		},
	})
}
