package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Async 协程异步处理，附带recover防止panic
func Async(f func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				zap.S().Error("Async Recover:", err)
			}
		}()
		f()
	}()
}

// PrintJson 将参数json格式化打印出来，方便观察、调试
func PrintJson(v interface{}) {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		zap.Error(err)
	}
	zap.S().WithOptions(zap.WithCaller(false)).Info(string(jsonData))
}

type CustomFileHeader struct {
	Filename    string
	FileBytes   []byte
	Size        int64
	OtherFields map[string]string
}

func CallApi(api, method string, timeout time.Duration, bodyParams, result interface{}) error {
	// 1. 准备请求数据
	var (
		requestBody []byte
		err         error
		contentType string
	)
	if fileHeader, ok := bodyParams.(*CustomFileHeader); ok {
		// 处理文件上传
		// 创建multipart writer
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		// 创建文件字段
		part, err := writer.CreateFormFile("file", fileHeader.Filename)
		if err != nil {
			return err
		}
		// 复制文件内容
		if _, err = io.Copy(part, bytes.NewReader(fileHeader.FileBytes)); err != nil {
			return err
		}
		// 写入其他字段
		for k, v := range fileHeader.OtherFields {
			if err = writer.WriteField(k, v); err != nil {
				return err
			}
		}
		// 关闭writer以完成multipart写入
		writer.Close()
		contentType = writer.FormDataContentType()
		requestBody = body.Bytes()
	} else {
		requestBody, err = json.Marshal(bodyParams)
		if err != nil {
			return err
		}
		contentType = "application/json"
	}

	// 2. 创建HTTP请求
	client := &http.Client{
		Timeout: timeout,
	}
	httpReq, err := http.NewRequest(method, api, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	// 3. 设置请求头
	httpReq.Header.Set("Content-Type", contentType)
	httpReq.Header.Set("Accept", "application/json")
	// 4. 发送请求
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// 5. 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// 6. 检查状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("CoreServer error: %s %s", resp.Status, string(body))
	}
	if err = json.Unmarshal(body, &result); err != nil {
		return err
	}
	return nil
}
