package hermes

import (
	"bytes"
	"html/template"

	"github.com/defeng-hub/Go-Email-Template/pkg/html2text"
	"github.com/defeng-hub/Go-Email-Template/pkg/premailer"
	"github.com/defeng-hub/Go-Email-Template/pkg/sprig"

	"github.com/imdario/mergo"
	"github.com/russross/blackfriday/v2"
)

// Hermes is an 实例 of the hermes email 生成器
type Hermes struct {
	Theme              Theme
	TextDirection      TextDirection
	Product            Product
	DisableCSSInlining bool
}

// Theme 是接口实现，创建一个新的主题
type Theme interface {
	Name() string              // The name of the theme
	HTMLTemplate() string      // The golang template for HTML emails
	PlainTextTemplate() string // The golang templte for plain text emails (can be basic HTML)
}

// TextDirection 电子邮件中的文本描述
type TextDirection string

var templateFuncs = template.FuncMap{
	"url": func(s string) template.URL {
		return template.URL(s)
	},
}

// TDLeftToRight 从左到右的文本方向(默认)
const TDLeftToRight TextDirection = "ltr"

// TDRightToLeft 从右到左的文本方向
const TDRightToLeft TextDirection = "rtl"

// Product 代表贵公司产品(品牌)
// 出现在电子邮件的页眉和页脚
type Product struct {
	Name        string
	Link        string // e.g. https://www.facec.cc
	Logo        string // e.g. https://matcornic.github.io/img/logo.png
	Copyright   string // Copyright © 2019 Hermes. All rights reserved.
	TroubleText string // TroubleText is the sentence at the end of the email for users having trouble with the button (default to `If you’re having trouble with the button '{ACTION}', copy and paste the URL below into your web browser.`)
}

// Email is the email containing a body
type Email struct {
	Body Body
}

// Markdown 是HTML模板(字符串)代表Markdown内容
// https://en.wikipedia.org/wiki/Markdown
type Markdown template.HTML

// Body is the body of the email, containing all interesting data
type Body struct {
	Name         string   // 联系人姓名
	Intros       []string // 介绍句，首先显示在电子邮件中
	Dictionary   []Entry  // 键+值列表(用于显示参数/设置/个人信息)
	Table        Table    // Table是一个可以放置数据(定价网格、账单等)的表。
	Actions      []Action // 操作是用户可以通过单击按钮执行的操作列表
	Outros       []string // 外句子，最后显示在电子邮件中
	Greeting     string   // 联系人的问候语(默认为'Hi')
	Signature    string   // 联系人签名(默认为“Yours truly”)
	Title        string   // Title 在设置时会替换问候语+名称
	FreeMarkdown Markdown // Free markdown content  替换除页眉和页脚以外的所有内容
}

// ToHTML 转换Markdown到HTML
func (c Markdown) ToHTML() template.HTML {
	return template.HTML(blackfriday.Run([]byte(string(c))))
}

// Entry 是一个简单的映射条目
// 允许使用entry的slice而不是map
// 因为Golang map没有排序
type Entry struct {
	Key   string
	Value string
}

// Table 是一个可以放置数据(定价网格、账单等)的表。
type Table struct {
	Data    [][]Entry // Contains data
	Columns Columns   // Contains meta-data for display purpose (width, alignement)
}

// Columns 包含不同列的元数据
type Columns struct {
	CustomWidth     map[string]string
	CustomAlignment map[string]string
}

// Action 是否有任何用户可以操作的内容(例如，点击按钮，查看邀请代码)
type Action struct {
	Instructions string
	Button       Button
	InviteCode   string
}

// 按钮定义一个要启动的动作
type Button struct {
	Color     string
	TextColor string
	Text      string
	Link      string
}

type Template struct {
	Hermes Hermes
	Email  Email
}

func SetDefaultEmailValues(e *Email) error {
	// 邮件默认值
	defaultEmail := Email{
		Body: Body{
			Intros:     []string{},
			Dictionary: []Entry{},
			Outros:     []string{},
			Signature:  "Thanks",
			Greeting:   "Hi",
		},
	}
	//将指定邮件与默认邮件合并
	//默认值覆盖所有零值
	return mergo.Merge(e, defaultEmail)
}

// SetDefaultHermesValues 引擎默认值
func SetDefaultHermesValues(h *Hermes) error {
	defaultTextDirection := TDLeftToRight
	defaultHermes := Hermes{
		Theme:         new(Default),
		TextDirection: defaultTextDirection,
		Product: Product{
			Name:        "Defeng",
			Copyright:   "Copyright © 2022 defeng-hub. All rights reserved.",
			TroubleText: "如果你无法点击 '{ACTION}' 按钮，请复制下面的URL到你的浏览器中访问",
		},
	}
	// 将给定的hermes引擎配置与默认配置合并
	// 默认值覆盖所有零值
	err := mergo.Merge(h, defaultHermes)
	if err != nil {
		return err
	}
	if h.TextDirection != TDLeftToRight && h.TextDirection != TDRightToLeft {
		h.TextDirection = defaultTextDirection
	}
	return nil
}

// GenerateHTML 生成从数据到HTML阅读器的电子邮件正文
// 这是为 现代电子邮件客户端
func (h *Hermes) GenerateHTML(email Email) (string, error) {
	err := SetDefaultHermesValues(h)
	if err != nil {
		return "", err
	}
	return h.generateTemplate(email, h.Theme.HTMLTemplate())
}

// GeneratePlainText 根据数据生成邮件正文
// 这是为 旧的电子邮件客户端
func (h *Hermes) GeneratePlainText(email Email) (string, error) {
	err := SetDefaultHermesValues(h)
	if err != nil {
		return "", err
	}
	template, err := h.generateTemplate(email, h.Theme.PlainTextTemplate())
	if err != nil {
		return "", err
	}
	return html2text.FromString(template, html2text.Options{PrettyTables: true})
}

func (h *Hermes) generateTemplate(email Email, tplt string) (string, error) {

	err := SetDefaultEmailValues(&email)
	if err != nil {
		return "", err
	}

	// Generate the email from Golang template
	// Allow usage of simple function from sprig : https://github.com/Masterminds/sprig
	// 从Golang模板生成邮件
	// 允许使用sprig中的简单函数:https://github.com/Masterminds/sprig
	t, err := template.New("hermes").Funcs(sprig.FuncMap()).Funcs(templateFuncs).Funcs(template.FuncMap{
		"safe": func(s string) template.HTML { return template.HTML(s) }, // Used for keeping comments in generated template
	}).Parse(tplt)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	err = t.Execute(&b, Template{*h, email})
	if err != nil {
		return "", err
	}

	res := b.String()
	if h.DisableCSSInlining {
		return res, nil
	}

	// 内联CSS
	prem, err := premailer.NewPremailerFromString(res, premailer.NewOptions())
	if err != nil {
		return "", err
	}
	html, err := prem.Transform()
	if err != nil {
		return "", err
	}
	return html, nil
}
