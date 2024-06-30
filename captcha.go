package captcha

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"os"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/google/uuid"
	"golang.org/x/image/font"
)

const (
	randString   = "ABCDEFGHIJKLMNOPQRSTUVWXYZadcdefghijklmnopqrstuvwxyz1234567890"
	fontFilePath = "./font/Arcade_Book.ttf"
	noisetext    = ".*^-~"
	W            = 120
	H            = 80
)

type Captcha struct {
	Cid       string
	Bs64Data  string
	Code      string
	Len       int
	StartTime time.Time
	Store     Storer
}

func NewCaptcha(codelen int) *Captcha {
	cid := uuid.New()
	return &Captcha{
		Cid: cid.String(),
		Len: codelen,
	}
}

func (c *Captcha)SetStoreMode(s Storer){
	c.Store = s
}

// getRandString 获取随机字符串
func (c *Captcha) getRandString(stringLen int) string {
	var result []byte
	rangeLen := len(randString)
	for i := 0; i < stringLen; i++ {
		result = append(result, randString[rand.Intn(rangeLen-1)])
	}
	return string(result)
}

func (c *Captcha) setCode() {
	randStr := c.getRandString(c.Len)
	c.Code = randStr
}

func (c *Captcha) GenerateCaptcha() (string, string, error) {
	c.setCode()
	ci := NewCaptchaImage(W, H, c.Len, 72.0)
	b64s, err := ci.drawImageText(c.Code)
	if err != nil {
		return "", "", err
	}
	//store the captcha
	err = c.Store.Save(c.Cid, *c)
	if err != nil {
		return "", "", err
	}
	return c.Cid, b64s, nil
}

// VerifyCaptcha 验证验证码是否一致
func (c *Captcha) VerifyCaptcha(id, code string) bool {
	if c.Store.Verify(id, code) {
		c.Store.Remove(id)
		return true
	}
	return false
}

type CaptchaImage struct {
	Width, Height, CodeLen int
	Fontsize               float64
	Dpi                    float64
	img                    *image.RGBA
}

func NewCaptchaImage(width, height, codelen int, dpi float64) *CaptchaImage {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	return &CaptchaImage{
		Width:   width,
		Height:  height,
		CodeLen: codelen,
		Dpi:     dpi,
		img:     img,
	}
}

// drawImageBGNoise 设置图片背景干扰字符
func (c *CaptchaImage) drawBGNoise(noiseText string) error {
	sc := freetype.NewContext()
	sc.SetDPI(c.Dpi)

	sc.SetClip(c.img.Bounds())
	sc.SetDst(c.img)
	sc.SetHinting(font.HintingFull)
	rawFontSize := float64(c.Height) / (1 + float64(rand.Intn(7))/float64(10))
	font, err := c.setFonts()
	if err != nil {
		return err
	}

	for _, char := range noiseText {
		rw := rand.Intn(c.Width)
		rh := rand.Intn(c.Height)
		fontSize := rawFontSize/2 + float64(rand.Intn(5))
		sc.SetSrc(image.NewUniform(color.RGBA{120, 120, 0, 200}))
		sc.SetFontSize(fontSize)
		sc.SetFont(font)
		pt := freetype.Pt(rw, rh)
		if _, err := sc.DrawString(string(char), pt); err != nil {
			return fmt.Errorf("draw noise: %w", err)
		}
	}
	return nil
}

// drawImageText 将content字符内容写入图片
func (c *CaptchaImage) drawImageText(content string) (string, error) {
	//draw image
	bgColor := c.setRandColor(0, 255)
	draw.Draw(c.img, c.img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)
	err := c.drawBGNoise(noisetext)
	if err != nil {
		return "", err
	}
	// image font settings
	font, err := c.setFonts()
	if err != nil {
		return "", err
	}
	fontSize := 36
	sc := freetype.NewContext()
	sc.SetDPI(72)
	sc.SetFont(font)
	sc.SetFontSize(float64(fontSize))
	sc.SetClip(c.img.Bounds())
	sc.SetDst(c.img)
	sc.SetSrc(&image.Uniform{color.RGBA{20, 20, 0, 255}})

	point := freetype.Pt(10, 60)
	for _, ch := range content {
		_, err := sc.DrawString(string(ch), point)
		if err != nil {
			return "", err
		}
		point.X += sc.PointToFixed(float64(fontSize * 5 / 7))
	}
	// file, err := os.Create("test.png")
	// if err != nil {
	// 	return err
	// }
	// defer file.Close()
	imgBuffer := bytes.NewBuffer(nil)
	err = png.Encode(imgBuffer, c.img)
	if err != nil {
		return "", err
	}
	b64s := c.EncodeB64string(imgBuffer.Bytes())
	return b64s, nil
}

func (c *CaptchaImage) EncodeB64string(buf []byte) string {
	return fmt.Sprintf("data:%s;base64,%s", "image/png", base64.StdEncoding.EncodeToString(buf))
}

// setFonts sets the font
func (c *CaptchaImage) setFonts() (*truetype.Font, error) {
	fontFile, err := os.ReadFile(fontFilePath)
	if err != nil {
		return nil, err
	}
	font, err := freetype.ParseFont(fontFile)
	if err != nil {
		return nil, fmt.Errorf("truetype parse: %w", err)
	}
	return font, nil
}

// setRandColor 设置随机颜色
func (c *CaptchaImage) setRandColor(min, max int) color.RGBA {
	if min > max {
		min = 0
		max = 255
	}

	return color.RGBA{
		R: uint8(rand.Intn(max-min) + min),
		G: uint8(rand.Intn(max-min) + min),
		B: uint8(rand.Intn(max-min) + min),
		A: uint8(rand.Intn(max-min) + min),
	}

}
