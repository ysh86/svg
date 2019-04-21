package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

type Group struct {
	ID          string  `xml:"id,attr"`
	Fill        string  `xml:"fill,attr"`
	Transform   string  `xml:"transform,attr"`
	StrokeWidth float32 `xml:"stroke-width,attr"`
	Stroke      string  `xml:"stroke,attr"`

	Groups []Group `xml:"g"`
	Paths  []Path  `xml:"path"`
}

func (g Group) String() string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("g: id=%q, fill=%#v, transform=%#v, stroke-width=%f, stroke=%#v, groups=%d, paths=%d\n", g.ID, g.Fill, g.Transform, g.StrokeWidth, g.Stroke, len(g.Groups), len(g.Paths)))
	for _, gg := range g.Groups {
		out.WriteString("  ")
		out.WriteString("  ")
		out.WriteString(gg.String())
	}
	for _, p := range g.Paths {
		out.WriteString("  ")
		out.WriteString("  ")
		out.WriteString("  ")
		out.WriteString(p.String())
	}

	return out.String()
}

type Path struct {
	ID string `xml:"id,attr"`
	D  string `xml:"d,attr"`
}

func (p Path) String() string {
	return fmt.Sprintf("path: id=%q, d=%#v", p.ID, p.D)
}

type Box struct {
	X      int
	Y      int
	Width  int
	Height int
}

type SvgRoot struct {
	XMLName xml.Name `xml:"svg"`

	ID      string `xml:"id,attr"`
	ViewBox string `xml:"viewBox,attr"` // Box
	Version string `xml:"version,attr"`

	Groups []Group `xml:"g"`
}

func (s *SvgRoot) String() string {
	return fmt.Sprintf("svg: id=%q, viewBox=%#v, version=%q, groups=%d", s.ID, s.ViewBox, s.Version, len(s.Groups))
}

func printIdent(level int) {
	for ; level > 0; level-- {
		fmt.Print("  ")
	}
}

func parseSVG(dec *xml.Decoder) (svg *SvgRoot, err error) {
	curTag := ""
	for {
		token, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local != "svg" {
				goto nextToken
			}
			curTag = t.Name.Local
			svg = &SvgRoot{XMLName: t.Name}
			for _, a := range t.Attr {
				switch a.Name.Local {
				case "id":
					svg.ID = a.Value
				case "viewBox":
					svg.ViewBox = a.Value
				case "version":
					svg.Version = a.Value
				}
			}
		case xml.EndElement:
			if t.Name.Local != curTag {
				return nil, fmt.Errorf("Unknown tag: %q", t.Name.Local)
			}
			curTag = ""
			goto nextToken
		case xml.ProcInst:
			// nothing to do
		case xml.CharData:
			// nothing to do
		default:
			return nil, fmt.Errorf("Error token: %#v", token)
		}
	}

nextToken:
	// TODO: parse

	return
}

func main() {
	dec := xml.NewDecoder(os.Stdin)

	svg, err := parseSVG(dec)
	if err != nil {
		panic(err)
	}

	fmt.Println(svg)
	panic(nil)

	level := 0
	for {
		token, err := dec.Token()
		if err == nil {
			switch t := token.(type) {
			case xml.StartElement:
				printIdent(level)
				fmt.Printf("%q:", t.Name.Local)
				for _, a := range t.Attr {
					fmt.Printf(" [%q, %q]", a.Name.Local, a.Value)
				}
				fmt.Println()
				level++
			case xml.EndElement:
				level--
			case xml.ProcInst:
				// nothing to do
			case xml.CharData:
				// nothing to do
			default:
				printIdent(level)
				fmt.Printf("Unknown: %#v\n", t)
			}
		}
		if err == io.EOF {
			printIdent(level)
			fmt.Println("EOF")
			break
		}
	}
}
