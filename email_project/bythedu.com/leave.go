package main

import hermes "github.com/defeng-hub/Go-Email-Template"

type Leave struct {
}

func (w *Leave) Name() string {
	return "leave"
}

func (w *Leave) Email() hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Name: "",
			Intros: []string{
				"您的请假申请已经提交，后续的最新消息我们将用邮件通知您。",
			},
			Actions: []hermes.Action{
				{
					Instructions: "点击按钮查看本次请假详情",
					Button: hermes.Button{
						Text: "查看详情",
						Link: "https://www.baidu.com/",
					},
				},
			},
			Outros: []string{
				"需要帮助，或者存在其他问题? 只要回复这封邮件，我们很乐意帮忙。",
			},
		},
	}
}
