package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"

	_ "image/jpeg"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type ShotInfo struct {
	Model        string
	Lens         string
	Focal        int64
	Aperture     string
	ShutterSpeed string
	ISO          string
}

func shutterSpeed(apex int) string {
	standards := map[int]string{
		-5: "30",
		-4: "15",
		-3: "8",
		-2: "4",
		-1: "2",
		0:  "1",
		1:  "1/2",
		2:  "1/4",
		3:  "1/8",
		4:  "1/15",
		5:  "1/30",
		6:  "1/60",
		7:  "1/125",
		8:  "1/250",
		9:  "1/500",
		10: "1/1000",
		11: "1/2000",
		12: "1/4000",
		13: "1/8000",
	}
	return standards[apex]
}

func aperture(apex int) string {
	return fmt.Sprintf("%v", math.Exp2(float64(apex)/2))
}

func (si *ShotInfo) String() string {
	return fmt.Sprintf("%s, %s, %vmm, F%s, %s, ISO %s", si.Model, si.Lens, si.Focal, si.Aperture, si.ShutterSpeed, si.ISO)
}

func decode(filename string) (*ShotInfo, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	exif.RegisterParsers(mknote.All...)
	x, err := exif.Decode(f)
	if err != nil {
		return nil, err
	}
	model, _ := x.Get(exif.Model)
	modelString, err := model.StringVal()
	if err != nil {
		return nil, err
	}
	ssv, _ := x.Get(exif.ShutterSpeedValue)
	srat, _ := ssv.Rat(0)
	if err != nil {
		return nil, err
	}
	sapexFloat, _ := srat.Float64()
	sapex := int(math.Round(sapexFloat))

	av, _ := x.Get(exif.ApertureValue)
	arat, _ := av.Rat(0)
	if err != nil {
		return nil, err
	}
	aapexFloat, _ := arat.Float64()
	aapex := int(math.Round(aapexFloat))

	focal, _ := x.Get(exif.FocalLength)
	numer, denom, _ := focal.Rat2(0)
	lens, _ := x.Get(exif.LensModel)
	lensString, err := lens.StringVal()
	if err != nil {
		return nil, err
	}

	iso, _ := x.Get(exif.ISOSpeedRatings)

	return &ShotInfo{
		Model:        modelString,
		Lens:         lensString,
		Focal:        numer / denom,
		ShutterSpeed: shutterSpeed(sapex),
		Aperture:     aperture(aapex),
		ISO:          iso.String(),
	}, nil
}

func bottomAdd(in string, out string, s string) error {
	inFile, err := os.Open(in)
	if err != nil {
		return err
	}
	defer inFile.Close()

	inImage, _, err := image.Decode(inFile)
	if err != nil {
		return err
	}

	outImage := image.NewRGBA(inImage.Bounds())

	point := fixed.Point26_6{X: fixed.I(0), Y: fixed.I(0)}
	d := &font.Drawer{
		Dst:  outImage,
		Src:  image.NewUniform(color.RGBA{200, 100, 0, 255}),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(s)
	return nil
}

func main() {
	source := "/mnt/local/photos/tarugasawafav/DSC_6846.jpg"
	out := "./DSC_6846_labeled.jpg"
	info, err := decode(source)
	if err != nil {
		log.Panic(err)
	}
	log.Println(info.String())
	log.Println(bottomAdd(source, out, info.String()))
}
