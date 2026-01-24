package pdf

import (
	"image"
	"time"

	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/webassembly"
)

var pool pdfium.Pool

func init() {
	var err error
	pool, err = webassembly.Init(webassembly.Config{
		MinIdle:  1,
		MaxIdle:  1,
		MaxTotal: 1,
	})
	if err != nil {
		panic(err)
	}
}

// ConvertToImages 將 PDF 每頁轉為高解析度圖片
func ConvertToImages(pdfPath string) ([]image.Image, error) {
	instance, err := pool.GetInstance(time.Second * 30)
	if err != nil {
		return nil, err
	}
	defer instance.Close()

	// 開啟 PDF
	doc, err := instance.OpenDocument(&requests.OpenDocument{
		FilePath: &pdfPath,
	})
	if err != nil {
		return nil, err
	}
	defer instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{
		Document: doc.Document,
	})

	// 取得頁數
	pageCount, err := instance.FPDF_GetPageCount(&requests.FPDF_GetPageCount{
		Document: doc.Document,
	})
	if err != nil {
		return nil, err
	}

	var images []image.Image
	for i := 0; i < pageCount.PageCount; i++ {
		// 渲染每頁為圖片 (300 DPI 高解析度)
		render, err := instance.RenderPageInDPI(&requests.RenderPageInDPI{
			DPI: 300,
			Page: requests.Page{
				ByIndex: &requests.PageByIndex{
					Document: doc.Document,
					Index:    i,
				},
			},
		})
		if err != nil {
			return nil, err
		}
		images = append(images, render.Result.Image)
	}

	return images, nil
}
