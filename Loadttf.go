package main

import (
	"io/ioutil"
	"log"

	"github.com/golang/freetype/truetype"
)


var mainFont *truetype.Font

func LoadFont(ttfPath string) error {
	data, err := ioutil.ReadFile(ttfPath)
	if err != nil {
			return err
	}
	font, err := truetype.Parse(data)
	if err != nil {
			return err
	}
	mainFont = font
	log.Println("[Font] Loaded font from:", ttfPath)
	return nil
}
