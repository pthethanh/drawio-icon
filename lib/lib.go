package lib

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JoshVarga/svgparser"
)

func Generate(outputFile, inputDir string) error {
	fs, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}
	lib, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer lib.Close()
	lib.WriteString("<mxlibrary>[")
	for i, f := range fs {
		t, err := os.ReadFile(filepath.Join(inputDir, f.Name()))
		if err != nil {
			return err
		}
		svg, err := svgparser.Parse(bytes.NewReader(t), false)
		if err != nil {
			return err
		}

		w, err := newWriter()
		if err != nil {
			return err
		}
		w.Process(svg)
		lib.WriteString(fmt.Sprintf(`{"title":"%s","data":"data:image/svg+xml;base64,%s;editableCssRules=.*;","w":24,"h":24,"aspect":"fixed"}`, f.Name(), w.Base64()))
		if i < len(fs)-1 {
			lib.WriteString(",")
		}
	}
	lib.WriteString("]</mxlibrary>")

	return nil
}

type writer struct {
	out       *bytes.Buffer
	outStyles []string
}

func newWriter() (*writer, error) {
	return &writer{
		out:       bytes.NewBuffer(nil),
		outStyles: make([]string, 0),
	}, nil
}

func (w *writer) Base64() string {
	return base64.StdEncoding.EncodeToString(w.out.Bytes())
}

func (w *writer) Process(e *svgparser.Element) {
	attrs := new(strings.Builder)
	styleName := fmt.Sprintf("style%d", len(w.outStyles))
	style := new(strings.Builder)
	style.WriteString(fmt.Sprintf(`.style%d{`, len(w.outStyles)))
	hasStyles := false

	for k, v := range e.Attributes {
		if k == "fill" || k == "stroke" || k == "color" {
			style.WriteString(fmt.Sprintf("%s:%s; ", k, v))
			hasStyles = true
			continue
		}
		// ignore class attribute as svg doesn't allow class to be redefined.
		if k == "class" {
			continue
		}
		attrs.WriteString(fmt.Sprintf(`%v=%q `, k, v))
	}
	if hasStyles {
		style.WriteString("}")
		w.outStyles = append(w.outStyles, style.String())

		attrs.WriteString(fmt.Sprintf(`class=%q `, styleName))
	}
	if e.Name == "path" {
		w.out.Write([]byte(fmt.Sprintf("<%s %s />", e.Name, attrs)))
	} else {
		w.out.Write([]byte(fmt.Sprintf("<%s %s>", e.Name, attrs)))
	}
	for _, c := range e.Children {
		w.Process(c)
	}
	if e.Name == "svg" && len(w.outStyles) > 0 {
		w.out.Write([]byte("<style>"))
		for _, cc := range w.outStyles {
			w.out.Write([]byte(cc))
		}
		w.out.Write([]byte("</style>"))
	}
	if e.Name != "path" {
		w.out.Write([]byte(fmt.Sprintf("</%s>", e.Name)))
	}
}
