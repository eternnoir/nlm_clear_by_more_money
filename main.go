package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"nlm_clear_by_more_money/gemini"
	"nlm_clear_by_more_money/pdf"
	"nlm_clear_by_more_money/pptx"
)

func main() {
	// 定義 CLI 參數
	sizeFlag := flag.String("size", "", "Image size (1K, 2K, 4K)")
	promptFlag := flag.String("prompt", "", "Enhancement prompt for Gemini")
	tempFlag := flag.Float64("temp", 0, "Temperature for generation (0.0-1.0)")
	parallelFlag := flag.Int("parallel", 5, "Number of parallel Gemini requests")
	noEnhance := flag.Bool("no-enhance", false, "Skip Gemini enhancement (for testing)")
	flag.Parse()

	// 取得 PDF 路徑（非 flag 參數）
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: nlm_clear_by_more_money [options] <input.pdf>")
		fmt.Println("\nOptions:")
		fmt.Println("  -size string    Image size (default: 2K, or env IMAGE_SIZE)")
		fmt.Println("  -prompt string  Enhancement prompt (default: built-in, or env ENHANCE_PROMPT)")
		fmt.Println("  -temp float     Temperature 0.0-1.0 (default: 0.6)")
		fmt.Println("  -parallel int   Number of parallel Gemini requests (default: 5)")
		fmt.Println("  -no-enhance     Skip Gemini enhancement (for testing)")
		fmt.Println("\nEnvironment variables:")
		fmt.Println("  GEMINI_API_KEY  (required unless -no-enhance) Gemini API key")
		fmt.Println("  IMAGE_SIZE      Default image size")
		fmt.Println("  ENHANCE_PROMPT  Default enhancement prompt")
		os.Exit(1)
	}
	inputPath := args[0]

	// 檢查 API Key (除非跳過增強)
	if !*noEnhance && os.Getenv("GEMINI_API_KEY") == "" {
		fmt.Println("Error: GEMINI_API_KEY environment variable not set")
		fmt.Println("Use -no-enhance to skip Gemini enhancement for testing")
		os.Exit(1)
	}

	// 建立配置：CLI 參數 > 環境變數 > 預設值
	config := gemini.DefaultConfig()

	// ImageSize: CLI > ENV > Default
	if *sizeFlag != "" {
		config.ImageSize = *sizeFlag
	} else if envSize := os.Getenv("IMAGE_SIZE"); envSize != "" {
		config.ImageSize = envSize
	}

	// Prompt: CLI > ENV > Default
	if *promptFlag != "" {
		config.Prompt = *promptFlag
	} else if envPrompt := os.Getenv("ENHANCE_PROMPT"); envPrompt != "" {
		config.Prompt = envPrompt
	}

	// Temperature: CLI > Default (no ENV for simplicity)
	if *tempFlag > 0 {
		config.Temperature = float32(*tempFlag)
	}

	// Parallel: 確保至少為 1
	parallel := *parallelFlag
	if parallel < 1 {
		parallel = 1
	}

	ctx := context.Background()

	// 顯示配置
	if !*noEnhance {
		fmt.Printf("Configuration:\n")
		fmt.Printf("  ImageSize: %s\n", config.ImageSize)
		fmt.Printf("  Prompt: %s\n", config.Prompt)
		fmt.Printf("  Temperature: %.2f\n", config.Temperature)
		fmt.Printf("  Parallel: %d\n", parallel)
		fmt.Println()
	} else {
		fmt.Println("Mode: No enhancement (testing)")
		fmt.Println()
	}

	// 1. PDF 轉圖片
	fmt.Println("Converting PDF to images...")
	images, err := pdf.ConvertToImages(inputPath)
	if err != nil {
		fmt.Printf("Error converting PDF: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Found %d pages\n", len(images))

	var enhancedImages [][]byte

	if *noEnhance {
		// 跳過 Gemini，直接使用原圖
		fmt.Println("Skipping Gemini enhancement...")
		for i, img := range images {
			var buf bytes.Buffer
			png.Encode(&buf, img)
			enhancedImages = append(enhancedImages, buf.Bytes())
			fmt.Printf("  Converted page %d/%d to PNG\n", i+1, len(images))
		}
	} else {
		// 2. 初始化 Gemini 客戶端
		client, err := gemini.NewClient(ctx, config)
		if err != nil {
			fmt.Printf("Error creating Gemini client: %v\n", err)
			os.Exit(1)
		}

		// 3. Gemini 增強每張圖片（並行處理）
		fmt.Printf("Enhancing images with Gemini (%s, %d parallel)...\n", config.ImageSize, parallel)
		enhancedImages = processImagesParallel(ctx, client, images, parallel)
	}

	// 4. 建立 PPTX
	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".pptx"
	fmt.Printf("Creating PPTX: %s\n", outputPath)
	if err := pptx.CreatePPTX(enhancedImages, outputPath); err != nil {
		fmt.Printf("Error creating PPTX: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done!")
}

// processImagesParallel 並行處理圖片增強
func processImagesParallel(ctx context.Context, client *gemini.Client, images []image.Image, parallel int) [][]byte {
	total := len(images)
	results := make([][]byte, total)

	// 使用 semaphore 控制並行數量
	sem := make(chan struct{}, parallel)
	var wg sync.WaitGroup
	var mu sync.Mutex
	completed := 0

	for i, img := range images {
		wg.Add(1)
		go func(idx int, img image.Image) {
			defer wg.Done()

			// 取得 semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// 處理圖片
			enhanced, err := client.EnhanceImage(ctx, img)
			if err != nil {
				// 失敗時使用原圖
				var buf bytes.Buffer
				png.Encode(&buf, img)
				results[idx] = buf.Bytes()

				mu.Lock()
				completed++
				fmt.Printf("  [%d/%d] Page %d: failed (%v), using original\n", completed, total, idx+1, err)
				mu.Unlock()
			} else {
				results[idx] = enhanced

				mu.Lock()
				completed++
				fmt.Printf("  [%d/%d] Page %d: done\n", completed, total, idx+1)
				mu.Unlock()
			}
		}(i, img)
	}

	wg.Wait()
	return results
}
