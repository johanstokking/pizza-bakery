package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"net/http"

	"github.com/johanstokking/pizza-bakery/internal"
)

const generateImageURL = "https://api.openai.com/v1/images/generations"

type imageGenerator struct {
	apiKey string
}

// NewImageGenerator creates a new ImageGenerator using DALL-E 3 from OpenAI.
func NewImageGenerator(apiKey string) internal.ImageGenerator {
	return &imageGenerator{apiKey: apiKey}
}

// Generate implements internal.ImageGenerator.
func (i *imageGenerator) Generate(ctx context.Context, description string) (image.Image, error) {
	req := struct {
		Prompt         string `json:"prompt"`
		Model          string `json:"model"`
		ResponseFormat string `json:"response_format"`
	}{
		Prompt:         description,
		Model:          "dall-e-3",
		ResponseFormat: "b64_json",
	}
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("openai: marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", generateImageURL, bytes.NewBuffer(reqData))
	if err != nil {
		return nil, fmt.Errorf("openai: prepare image request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+i.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	log.Println("openai: generating image with prompt:", description)
	httpRes, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai: perform image request: %w", err)
	}
	defer httpRes.Body.Close()
	resData, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, fmt.Errorf("openai: perform image request: %w", err)
	}
	if httpRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai: unexpected status code: %d", httpRes.StatusCode)
	}
	res := struct {
		Data []struct {
			B64JSON string `json:"b64_json"`
		} `json:"data"`
	}{}
	if err := json.Unmarshal(resData, &res); err != nil {
		return nil, fmt.Errorf("openai: unmarshal response: %w", err)
	}

	if len(res.Data) == 0 {
		return nil, fmt.Errorf("openai: no image data in response")
	}
	dec := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(res.Data[0].B64JSON))
	img, err := png.Decode(dec)
	if err != nil {
		return nil, fmt.Errorf("openai: decode image: %w", err)
	}
	return img, nil
}
