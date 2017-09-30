package main

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"strings"
    "math/rand"
    "regexp"
    "fmt"
    
    "github.com/anthonynsimon/bild/transform"
    "github.com/anthonynsimon/bild/effect"
    "github.com/anthonynsimon/bild/blur"
    "github.com/anthonynsimon/bild/blend"
    "github.com/anthonynsimon/bild/noise"
    "github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var (
	fontDpi     = 72.0         // font DPI setting
	fontHinting = "full"       // none | full
	fontSize    = 72.0         // font size in points
)

// Load Font from disk
func loadFont(filename string) *truetype.Font {
    // Read the font data
    fontBytes, err := ioutil.ReadFile(RESOURCES_FOLDER + filename)
    if err != nil {
        log.Fatalln("ERROR, Loading font:", err)
    }

    fontTtf, err := truetype.Parse(fontBytes)
    if err != nil {
        log.Fatalln("ERROR, Parsing font: %s FileName: %s", err, filename)
    }
    return fontTtf
}

func isValidFontFileName(filename string) bool {
    filenamelower := strings.ToLower(filename)  
    matched, _ := regexp.MatchString(".*.tt[fc]",filenamelower) 
    return matched
}

func loadRandomFont() *truetype.Font {
    files, err := ioutil.ReadDir(RESOURCES_FOLDER) 
    if err != nil {
        log.Fatalln("ERROR, Reading Directory:", err)
    }
    fontfiles := []string{}
    for _,file := range files {
        if isValidFontFileName(file.Name()){
            fontfiles = append(fontfiles, file.Name())
        }
    }

    n := rand.Int() % len(fontfiles)
    return loadFont(fontfiles[n])
}

func AddImageEffects(input *image.RGBA) *image.RGBA {
    rgba := input
    //debug string
    debugParams := "Effects debug: "

    // Add random transform
    rgba = transform.Rotate(rgba, 10 * (0.5 - rand.Float64()), nil)
    rgba = transform.ShearV(rgba, 5 * (0.5 - rand.Float64())) 
    rgba = transform.ShearH(rgba, 5 * (0.5 - rand.Float64()))

    // Create noise
    width := rgba.Bounds().Size().X
    height := rgba.Bounds().Size().Y
    isMonochrome := (rand.Int() % 2) == 0
    noisefn := noise.Gaussian
    switch (rand.Int() % 3) {
    case 0:
        noisefn = noise.Binary
        debugParams = debugParams + " Binary Noise"
    case 1:
        noisefn = noise.Uniform
        debugParams = debugParams + " Uniform Noise"
    default: 
        debugParams = debugParams + " Gaussian Noise"
    }
    noise := noise.Generate(width, height, &noise.Options{NoiseFn: noisefn, Monochrome: isMonochrome})

    // Blend noise
    switch (rand.Int() % 6) {
    case 0:
        rgba = blend.Opacity(rgba, noise, 0.5)
        debugParams = debugParams + " Opacity Blend"
    case 1:
        rgba = blend.Lighten(rgba, noise)
        debugParams = debugParams + " Lighten Blend"
    case 2:
        rgba = blend.Subtract(rgba, noise)
        debugParams = debugParams + " Subtraction Blend"
    case 3:
        rgba = blend.SoftLight(rgba, noise)
        debugParams = debugParams + " SoftLight Blend"
    case 4:
        rgba = blend.ColorBurn(rgba, noise)
        debugParams = debugParams + " ColorBurn Blend"
    case 5:
        rgba = blend.Overlay(rgba, noise)
        debugParams = debugParams + " Overlay Blend"
    default:
        rgba = blend.Exclusion(rgba, noise)
        debugParams = debugParams + " Exclusion Blend"
    }

    //Add effect
    switch (rand.Int() % 4) {
    case 0:
        rgba = blur.Gaussian(rgba, 2.0)
        debugParams = debugParams + " Gaussian Blur Effect"
    case 1:
        rgba = effect.Emboss(rgba)
        debugParams = debugParams + " Emboss Effect"
    case 2:
        rgba = blur.Box(rgba, 1.8)
        debugParams = debugParams + " Box Blur Effect"
    default:
        rgba = effect.Sobel(rgba)
        debugParams = debugParams + " Sobel Effect"
    }

    //apply mirrored challege mode effects
    switch (rand.Int() % 4) {
    case 0:
    //apply mirrored 
        rgba = transform.FlipH(rgba)
    case 1:
    //apply upsidedown
        rgba = transform.Rotate(rgba,180,nil)
    case 2:
    //apply upsidedown and mirrored
        rgba = transform.FlipH(rgba)
        rgba = transform.Rotate(rgba,180,nil)
    default:
    }
    //Debug String Print
    fmt.Println(debugParams)

    return rgba
}


// Generate a PNG image reader with given string written
func GenerateImage(input string, effects bool) *bytes.Buffer {

	if len(input) == 0 {
		log.Println("ERROR, Can't generate image without input")
		return nil
	}

	// Set up font hinting
	h := font.HintingNone
	switch fontHinting {
	case "full":
		h = font.HintingFull
	}

	// Pick colours
	fg, bg := image.Black, image.White

	// Set up font drawer
	d := &font.Drawer{
		Src: fg,
		Face: truetype.NewFace(loadRandomFont(), &truetype.Options{
			Size:    fontSize,
			DPI:     fontDpi,
			Hinting: h,
		}),
	}

	// Prepare lines to be drawn
	lines := strings.Split(input, "\n")

	// Figure out image bounds
	var widest int
	for _, line := range lines {
		width := d.MeasureString(line).Round()
		if width > widest {
			widest = width
		}
	}

	lineHeight := int(math.Ceil(fontSize * fontDpi / 72 * 1.18))
	imgW := widest * 11 / 10 // 10% extra for margins
	imgH := len(lines) * lineHeight

	// Create image canvas
	rgba := image.NewRGBA(image.Rect(0, 0, imgW, imgH))

	// Draw the background and the guidelines
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)

	// Attach image to font drawer
	d.Dst = rgba

	// Figure out writing position
	y := int(math.Ceil(fontSize * fontDpi / 72 * 0.94))
	x := fixed.I(imgW-widest) / 2
	for _, line := range lines {
		d.Dot = fixed.Point26_6{
			X: x,
			Y: fixed.I(y),
		}

		// Write out the text
		d.DrawString(line)

		// Advance line position
		y += lineHeight
	}
  
    if (effects) {
        rgba = AddImageEffects(rgba)
    }

	// Encode PNG image
	var buf bytes.Buffer
	err := png.Encode(&buf, rgba)
	if err != nil {
		log.Println("ERROR, Encoding PNG with '"+input+"':", err)
		return &buf
	}

	return &buf
}
