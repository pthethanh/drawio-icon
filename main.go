package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JoshVarga/svgparser"
)

func main() {
	in := flag.String("input", "icons", "input dir")
	out := flag.String("output", "out", "output dir")
	flag.Parse()

	fs, err := os.ReadDir(*in)
	if err != nil {
		panic(err)
	}
	lib, err := os.Create(filepath.Join(*out, "my_drawio_lib.xml"))
	if err != nil {
		panic(err)
	}
	lib.WriteString("<mxlibrary>[")
	for i, f := range fs {
		t, err := os.ReadFile(filepath.Join(*in, f.Name()))
		if err != nil {
			panic(err)
		}
		svg, err := svgparser.Parse(bytes.NewReader(t), false)
		if err != nil {
			panic(err)
		}

		w, err := newWriter()
		if err != nil {
			panic(err)
		}
		w.Process(svg)
		if err := os.MkdirAll(*out, 0644); err != nil {
			panic(err)
		}
		out, err := os.Create(filepath.Join(*out, f.Name()))
		if err != nil {
			panic(err)
		}
		out.Write(w.out.Bytes())
		out.Close()
		lib.WriteString(fmt.Sprintf(`{"title":"%s","data":"data:image/svg+xml;base64,%s;editableCssRules=.*;","w":24,"h":24,"aspect":"fixed"}`, f.Name(), w.Base64()))
		if i < len(fs)-1 {
			lib.WriteString(",")
		}
	}
	lib.WriteString("]</mxlibrary>")
	lib.Close()
}

type Writer struct {
	out       *bytes.Buffer
	outStyles []string
}

func newWriter() (*Writer, error) {
	return &Writer{
		out:       bytes.NewBuffer(nil),
		outStyles: make([]string, 0),
	}, nil
}

func (w *Writer) Base64() string {
	return base64.StdEncoding.EncodeToString(w.out.Bytes())
}

func (w *Writer) Process(e *svgparser.Element) {
	attrs := bytes.NewBuffer(nil)
	styleName := fmt.Sprintf("style%d", len(w.outStyles))
	style := fmt.Sprintf(`.style%d{`, len(w.outStyles))
	hasStyles := false
	for k, v := range e.Attributes {
		if k == "fill" || k == "stroke" || k == "color" {
			style += fmt.Sprintf("%s:%s; ", k, v)
			hasStyles = true
		} else {
			attrs.WriteString(fmt.Sprintf(`%v=%q `, k, v))
		}
	}
	if hasStyles {
		style += "}"
		attrs.WriteString(fmt.Sprintf(`class=%q `, styleName))

		w.outStyles = append(w.outStyles, style)

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
