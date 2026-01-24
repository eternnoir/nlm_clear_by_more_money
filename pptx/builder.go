package pptx

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

// Slide dimensions in EMU (English Metric Units)
// 1 inch = 914400 EMU, standard 16:9 slide is 10" x 5.625"
const (
	slideWidthEMU  = 9144000  // 10 inches
	slideHeightEMU = 5143500  // 5.625 inches (16:9)
)

// CreatePPTX 將圖片 bytes 陣列打包成 PPTX
func CreatePPTX(imageDataList [][]byte, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	// 1. [Content_Types].xml
	if err := writeFile(zw, "[Content_Types].xml", contentTypesXML(len(imageDataList))); err != nil {
		return err
	}

	// 2. _rels/.rels
	if err := writeFile(zw, "_rels/.rels", relsXML()); err != nil {
		return err
	}

	// 3. docProps/app.xml
	if err := writeFile(zw, "docProps/app.xml", appXML()); err != nil {
		return err
	}

	// 4. docProps/core.xml
	if err := writeFile(zw, "docProps/core.xml", coreXML()); err != nil {
		return err
	}

	// 5. ppt/presentation.xml
	if err := writeFile(zw, "ppt/presentation.xml", presentationXML(len(imageDataList))); err != nil {
		return err
	}

	// 6. ppt/_rels/presentation.xml.rels
	if err := writeFile(zw, "ppt/_rels/presentation.xml.rels", presentationRelsXML(len(imageDataList))); err != nil {
		return err
	}

	// 7. ppt/theme/theme1.xml
	if err := writeFile(zw, "ppt/theme/theme1.xml", themeXML()); err != nil {
		return err
	}

	// 8. ppt/slideMasters/slideMaster1.xml
	if err := writeFile(zw, "ppt/slideMasters/slideMaster1.xml", slideMasterXML()); err != nil {
		return err
	}

	// 9. ppt/slideMasters/_rels/slideMaster1.xml.rels
	if err := writeFile(zw, "ppt/slideMasters/_rels/slideMaster1.xml.rels", slideMasterRelsXML()); err != nil {
		return err
	}

	// 10. ppt/slideLayouts/slideLayout1.xml
	if err := writeFile(zw, "ppt/slideLayouts/slideLayout1.xml", slideLayoutXML()); err != nil {
		return err
	}

	// 11. ppt/slideLayouts/_rels/slideLayout1.xml.rels
	if err := writeFile(zw, "ppt/slideLayouts/_rels/slideLayout1.xml.rels", slideLayoutRelsXML()); err != nil {
		return err
	}

	// 12. Add slides and images
	for i, imgData := range imageDataList {
		slideNum := i + 1

		// Detect image dimensions
		imgWidth, imgHeight := getImageDimensions(imgData)

		// Calculate scaling to fit slide while maintaining aspect ratio
		width, height, offsetX, offsetY := calculateFitDimensions(imgWidth, imgHeight)

		// Write image file
		ext := detectImageType(imgData)
		imgPath := fmt.Sprintf("ppt/media/image%d.%s", slideNum, ext)
		if err := writeFileBytes(zw, imgPath, imgData); err != nil {
			return err
		}

		// Write slide
		slidePath := fmt.Sprintf("ppt/slides/slide%d.xml", slideNum)
		if err := writeFile(zw, slidePath, slideXML(slideNum, width, height, offsetX, offsetY)); err != nil {
			return err
		}

		// Write slide rels
		slideRelsPath := fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", slideNum)
		if err := writeFile(zw, slideRelsPath, slideRelsXML(slideNum, ext)); err != nil {
			return err
		}
	}

	return nil
}

func writeFile(zw *zip.Writer, name, content string) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(content))
	return err
}

func writeFileBytes(zw *zip.Writer, name string, data []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func detectImageType(data []byte) string {
	if len(data) > 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "jpeg"
	}
	return "png"
}

func getImageDimensions(data []byte) (int, int) {
	img, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 800, 600 // default fallback
	}
	return img.Width, img.Height
}

func calculateFitDimensions(imgWidth, imgHeight int) (width, height, offsetX, offsetY int64) {
	// Calculate scale to fit within slide
	scaleW := float64(slideWidthEMU) / float64(imgWidth)
	scaleH := float64(slideHeightEMU) / float64(imgHeight)

	scale := scaleW
	if scaleH < scaleW {
		scale = scaleH
	}

	width = int64(float64(imgWidth) * scale)
	height = int64(float64(imgHeight) * scale)

	// Center the image
	offsetX = (slideWidthEMU - width) / 2
	offsetY = (slideHeightEMU - height) / 2

	return
}

func contentTypesXML(slideCount int) string {
	slides := ""
	images := ""
	for i := 1; i <= slideCount; i++ {
		slides += fmt.Sprintf(`<Override PartName="/ppt/slides/slide%d.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>`, i)
	}
	for i := 1; i <= slideCount; i++ {
		images += fmt.Sprintf(`<Override PartName="/ppt/media/image%d.png" ContentType="image/png"/>`, i)
	}
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Default Extension="png" ContentType="image/png"/>
<Default Extension="jpeg" ContentType="image/jpeg"/>
<Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>
<Override PartName="/ppt/slideMasters/slideMaster1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideMaster+xml"/>
<Override PartName="/ppt/slideLayouts/slideLayout1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideLayout+xml"/>
<Override PartName="/ppt/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/>
<Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>
<Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>
` + slides + images + `
</Types>`
}

func relsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
<Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>
</Relationships>`
}

func appXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties">
<Application>nlm_clear_by_more_money</Application>
</Properties>`
}

func coreXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/">
<dc:creator>nlm_clear_by_more_money</dc:creator>
</cp:coreProperties>`
}

func presentationXML(slideCount int) string {
	slideList := ""
	for i := 1; i <= slideCount; i++ {
		slideList += fmt.Sprintf(`<p:sldId id="%d" r:id="rId%d"/>`, 255+i, i)
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:presentation xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" saveSubsetFonts="1">
<p:sldMasterIdLst><p:sldMasterId id="2147483648" r:id="rId%d"/></p:sldMasterIdLst>
<p:sldIdLst>%s</p:sldIdLst>
<p:sldSz cx="%d" cy="%d"/>
<p:notesSz cx="%d" cy="%d"/>
</p:presentation>`, slideCount+1, slideList, slideWidthEMU, slideHeightEMU, slideWidthEMU, slideHeightEMU)
}

func presentationRelsXML(slideCount int) string {
	rels := ""
	for i := 1; i <= slideCount; i++ {
		rels += fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide%d.xml"/>`, i, i)
	}
	masterID := slideCount + 1
	themeID := slideCount + 2
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
%s
<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="slideMasters/slideMaster1.xml"/>
<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="theme/theme1.xml"/>
</Relationships>`, rels, masterID, themeID)
}

func themeXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="Office Theme">
<a:themeElements>
<a:clrScheme name="Office"><a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1><a:lt1><a:sysClr val="window" lastClr="FFFFFF"/></a:lt1><a:dk2><a:srgbClr val="44546A"/></a:dk2><a:lt2><a:srgbClr val="E7E6E6"/></a:lt2><a:accent1><a:srgbClr val="4472C4"/></a:accent1><a:accent2><a:srgbClr val="ED7D31"/></a:accent2><a:accent3><a:srgbClr val="A5A5A5"/></a:accent3><a:accent4><a:srgbClr val="FFC000"/></a:accent4><a:accent5><a:srgbClr val="5B9BD5"/></a:accent5><a:accent6><a:srgbClr val="70AD47"/></a:accent6><a:hlink><a:srgbClr val="0563C1"/></a:hlink><a:folHlink><a:srgbClr val="954F72"/></a:folHlink></a:clrScheme>
<a:fontScheme name="Office"><a:majorFont><a:latin typeface="Calibri Light"/><a:ea typeface=""/><a:cs typeface=""/></a:majorFont><a:minorFont><a:latin typeface="Calibri"/><a:ea typeface=""/><a:cs typeface=""/></a:minorFont></a:fontScheme>
<a:fmtScheme name="Office"><a:fillStyleLst><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:fillStyleLst><a:lnStyleLst><a:ln w="6350"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln><a:ln w="12700"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln><a:ln w="19050"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln></a:lnStyleLst><a:effectStyleLst><a:effectStyle><a:effectLst/></a:effectStyle><a:effectStyle><a:effectLst/></a:effectStyle><a:effectStyle><a:effectLst/></a:effectStyle></a:effectStyleLst><a:bgFillStyleLst><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:bgFillStyleLst></a:fmtScheme>
</a:themeElements>
</a:theme>`
}

func slideMasterXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldMaster xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
<p:cSld><p:spTree><p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr><p:grpSpPr/></p:spTree></p:cSld>
<p:clrMap bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"/>
<p:sldLayoutIdLst><p:sldLayoutId id="2147483649" r:id="rId1"/></p:sldLayoutIdLst>
</p:sldMaster>`
}

func slideMasterRelsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="../theme/theme1.xml"/>
</Relationships>`
}

func slideLayoutXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldLayout xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" type="blank">
<p:cSld><p:spTree><p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr><p:grpSpPr/></p:spTree></p:cSld>
</p:sldLayout>`
}

func slideLayoutRelsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="../slideMasters/slideMaster1.xml"/>
</Relationships>`
}

func slideXML(slideNum int, width, height, offsetX, offsetY int64) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
<p:cSld>
<p:spTree>
<p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
<p:grpSpPr/>
<p:pic>
<p:nvPicPr><p:cNvPr id="2" name="Image %d"/><p:cNvPicPr><a:picLocks noChangeAspect="1"/></p:cNvPicPr><p:nvPr/></p:nvPicPr>
<p:blipFill><a:blip r:embed="rId1"/><a:stretch><a:fillRect/></a:stretch></p:blipFill>
<p:spPr>
<a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm>
<a:prstGeom prst="rect"><a:avLst/></a:prstGeom>
</p:spPr>
</p:pic>
</p:spTree>
</p:cSld>
</p:sld>`, slideNum, offsetX, offsetY, width, height)
}

func slideRelsXML(slideNum int, imgExt string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="../media/image%d.%s"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>
</Relationships>`, slideNum, imgExt)
}
