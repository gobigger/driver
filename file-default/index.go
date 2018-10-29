package file_default

import (
	"strconv"
	"time"
	. "github.com/gobigger/bigger"
)


const (
	browseName	= "*.$.default.browse"
	previewName	= "*.$.default.preview"
)

func Driver() (FileDriver) {
	return &defaultFileDriver{}
}

func init() {
	Bigger.Driver("default", Driver())

	//自带一个文件浏览和预览的路由
	Bigger.Router(browseName, Map{
		"uri": "/file/browse/{code}",
		"name": "浏览文件", "text": "浏览文件",
		"args": Map{
			"code": Map{
				"type": "string", "must": true, "name": "文件编码", "text": "文件编码",
			},
			"name": Map{
				"type": "string", "must": false, "name": "自定义文件名", "text": "自定义文件名",
			},
			"token": Map{
				"type": "[string]", "must": true, "name": "令牌", "text": "令牌",
				"encode": "strings", "decode": "strings",
			},
		},
		"action": func(ctx *Context) {

			code := ctx.Args["code"].(string)
			data := Bigger.Decode(code)
			if data == nil {
				ctx.Text("无效访问代码")
				return
			}

			tokens := ctx.Args["token"].([]string)
			if len(tokens) != 3 {
				ctx.Text("无效访问令牌1")
				return
			}

			ip := tokens[0]
			if ip != "" && ctx.Ip() != ip {
				ctx.Text("无效访问令牌2")
				return
			}

			expiry := int64(-1)
			if vv,ee := strconv.ParseInt(tokens[1], 10, 64); ee == nil {
				expiry = vv
			}
			if expiry >= 0 {
				if expiry < time.Now().UnixNano() {
					ctx.Text("无效访问令牌3")
					return
				}
			}
			if data.Name != tokens[2] {
				//超时不让访问了
				ctx.Text("无效访问令牌4")
				return
			}

			ctx.Download(code)
		},
	})


	//自带一个文件浏览和预览的路由
	Bigger.Router(previewName, Map{
		"uri": "/file/preview/{size}/{code}",
		"name": "预览文件", "text": "预览文件",
		"args": Map{
			"code": Map{
				"type": "string", "must": true, "name": "文件编码", "text": "文件编码",
			},
			"size": Map{
				"type": "[int]", "must": true, "name": "文件编码", "text": "文件编码",
				"encode": "numbers", "decode": "numbers",
			},
			"token": Map{
				"type": "[string]", "must": true, "name": "令牌", "text": "令牌",
				"encode": "strings", "decode": "strings",
			},
		},
		"action": func(ctx *Context) {
			
			code := ctx.Args["code"].(string)
			data := Bigger.Decode(code)
			if data == nil {
				ctx.Text("无效访问代码")
				return
			}

			tokens := ctx.Args["token"].([]string)
			if len(tokens) != 3 {
				ctx.Text("无效访问令牌1")
				return
			}

			ip := tokens[0]
			if ip != "" && ctx.Ip() != ip {
				ctx.Text("无效访问令牌2")
				return
			}

			expiry := int64(-1)
			if vv,ee := strconv.ParseInt(tokens[1], 10, 64); ee == nil {
				expiry = vv
			}
			if expiry >= 0 {
				if expiry < time.Now().UnixNano() {
					ctx.Text("无效访问令牌3")
					return
				}
			}
			if data.Name != tokens[2] {
				//超时不让访问了
				ctx.Text("无效访问令牌4")
				return
			}
			
			size := ctx.Args["size"].([]int64)
			if len(size) != 3 {
				ctx.Found()
				return
			}

			ctx.Thumbnail(code, size[0], size[1], size[2])
		},
	})
}


