package gemini

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"os"

	"google.golang.org/genai"
)

const defaultModel = "gemini-3.1-flash-image-preview"

func getModel() string {
	if m := os.Getenv("GEMINI_MODEL"); m != "" {
		return m
	}
	return defaultModel
}

// Config 儲存 Gemini API 的配置參數
type Config struct {
	ImageSize   string  // 圖片解析度: "1K", "2K", "4K"
	Prompt      string  // System instruction prompt
	Temperature float32 // 生成溫度
}

// DefaultConfig 返回預設配置
func DefaultConfig() Config {
	return Config{
		ImageSize:   "2K",
		Prompt:      "Carefully repair and enhance the blurry or unclear text and images in this picture",
		Temperature: 0.6,
	}
}

type Client struct {
	client *genai.Client
	config Config
}

func NewClient(ctx context.Context, config Config) (*Client, error) {
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Client{client: client, config: config}, nil
}

// EnhanceImage 使用 Gemini 增強圖片
func (c *Client) EnhanceImage(ctx context.Context, img image.Image) ([]byte, error) {
	// 圖片轉 PNG bytes
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	// 建構請求內容
	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				genai.NewPartFromBytes(buf.Bytes(), "image/png"),
				genai.NewPartFromText("請修復並增強這張圖片"),
			},
		},
	}

	// 設定生成參數（使用配置）
	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr[float32](c.config.Temperature),
		ResponseModalities: []string{
			"IMAGE",
			"TEXT",
		},
		ImageConfig: &genai.ImageConfig{
			ImageSize: c.config.ImageSize,
		},
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				genai.NewPartFromText(c.config.Prompt),
			},
		},
	}

	// 使用 streaming 取得結果
	var resultData []byte
	for result, err := range c.client.Models.GenerateContentStream(ctx, getModel(), contents, config) {
		if err != nil {
			return nil, err
		}

		if len(result.Candidates) == 0 || result.Candidates[0].Content == nil {
			continue
		}

		for _, part := range result.Candidates[0].Content.Parts {
			if part.InlineData != nil {
				resultData = part.InlineData.Data
			}
		}
	}

	if resultData == nil {
		return nil, fmt.Errorf("no image returned from Gemini")
	}

	return resultData, nil
}
