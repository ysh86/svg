package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Color struct {
	R int
	G int
	B int
	A int
}

func (c *Color) parse(s string) {
	//if s == "none" {
	c.R = 0
	c.G = 0
	c.B = 0
	c.A = 0
	//}

	// hex
	if strings.HasPrefix(s, "#") {
		switch len(s) {
		case 1 + 3:
			r, _ := strconv.ParseUint(s[1+0:1+1], 16, 8)
			g, _ := strconv.ParseUint(s[1+1:1+2], 16, 8)
			b, _ := strconv.ParseUint(s[1+2:1+3], 16, 8)
			a := 255
			c.R = int((r << 4) | r)
			c.G = int((g << 4) | g)
			c.B = int((b << 4) | b)
			c.A = int(a)
		case 1 + 6:
			r, _ := strconv.ParseUint(s[1+0:1+2], 16, 8)
			g, _ := strconv.ParseUint(s[1+2:1+4], 16, 8)
			b, _ := strconv.ParseUint(s[1+4:1+6], 16, 8)
			a := 255
			c.R = int(r)
			c.G = int(g)
			c.B = int(b)
			c.A = int(a)
		case 1 + 4:
			r, _ := strconv.ParseUint(s[1+0:1+1], 16, 8)
			b, _ := strconv.ParseUint(s[1+1:1+2], 16, 8)
			g, _ := strconv.ParseUint(s[1+2:1+3], 16, 8)
			a, _ := strconv.ParseUint(s[1+3:1+4], 16, 8)
			c.R = int((r << 4) | r)
			c.G = int((g << 4) | g)
			c.B = int((b << 4) | b)
			c.A = int((a << 4) | a)
		case 1 + 8:
			r, _ := strconv.ParseUint(s[1+0:1+2], 16, 8)
			g, _ := strconv.ParseUint(s[1+2:1+4], 16, 8)
			b, _ := strconv.ParseUint(s[1+4:1+6], 16, 8)
			a, _ := strconv.ParseUint(s[1+6:1+8], 16, 8)
			c.R = int(r)
			c.G = int(g)
			c.B = int(b)
			c.A = int(a)
		}
	}
}

// Matrix
// a c e   x   ax + cy + e
// b d f . y = bx + dy + f
// 0 0 1   1   0  + 0  + 1
type Matrix struct {
	a float32
	b float32
	c float32
	d float32
	e float32
	f float32
}

func (m *Matrix) parse(s string) {
	m.a = 1.0
	m.b = 0
	m.c = 0
	m.d = 1.0
	m.e = 0
	m.f = 0

	prefix := "matrix("
	if strings.HasPrefix(s, prefix) && strings.HasSuffix(s, ")") {
		ms := strings.Split(s[len(prefix):len(s)-1], ",")

		if len(ms) == 6 {
			a, _ := strconv.ParseFloat(ms[0], 32)
			b, _ := strconv.ParseFloat(ms[1], 32)
			c, _ := strconv.ParseFloat(ms[2], 32)
			d, _ := strconv.ParseFloat(ms[3], 32)
			e, _ := strconv.ParseFloat(ms[4], 32)
			f, _ := strconv.ParseFloat(ms[5], 32)
			m.a = float32(a)
			m.b = float32(b)
			m.c = float32(c)
			m.d = float32(d)
			m.e = float32(e)
			m.f = float32(f)
		}
	}
}

type Group struct {
	ID          string  `xml:"id,attr"`
	Fill        Color   `xml:"fill,attr"`
	Transform   Matrix  `xml:"transform,attr"`
	StrokeWidth float32 `xml:"stroke-width,attr"`
	Stroke      Color   `xml:"stroke,attr"`

	Groups []*Group `xml:"g"`
	Paths  []*Path  `xml:"path"`
}

func (g *Group) String() string {
	return fmt.Sprintf("g: id=%q, fill=%#v, transform=%#v, stroke-width=%f, stroke=%#v, groups=%d, paths=%d", g.ID, g.Fill, g.Transform, g.StrokeWidth, g.Stroke, len(g.Groups), len(g.Paths))
}

func (g *Group) parse(dec *xml.Decoder, token xml.Token) (xml.Token, error) {
	var err error

	curTag := ""
	for {
		switch t := token.(type) {
		case xml.StartElement:
			if curTag == "g" && t.Name.Local == "g" {
				g2 := new(Group)
				token, err = g2.parse(dec, token)
				if err != nil {
					return token, err
				}
				g.Groups = append(g.Groups, g2)
				continue
			}
			if curTag == "g" && t.Name.Local == "path" {
				p := new(Path)
				token, err = p.parse(dec, token)
				if err != nil {
					return token, err
				}
				g.Paths = append(g.Paths, p)
				continue
			}
			if curTag != "" || t.Name.Local != "g" {
				return token, fmt.Errorf("Invalid format")
			}
			curTag = t.Name.Local
			for _, a := range t.Attr {
				switch a.Name.Local {
				case "id":
					g.ID = a.Value
				case "fill":
					g.Fill.parse(a.Value)
				case "transform":
					g.Transform.parse(a.Value)
				case "stroke-width":
					f, _ := strconv.ParseFloat(a.Value, 32)
					g.StrokeWidth = float32(f)
				case "stroke":
					g.Stroke.parse(a.Value)
				}
			}
		case xml.EndElement:
			if curTag == "" {
				return token, fmt.Errorf("Invalid format")
			}
			if t.Name.Local != curTag {
				return token, fmt.Errorf("Unknown tag: %q", t.Name.Local)
			}
			curTag = ""
			return dec.Token()
		case xml.ProcInst:
			// nothing to do
		case xml.CharData:
			// nothing to do
		default:
			return token, fmt.Errorf("Error token: %#v", token)
		}

		token, err = dec.Token()
		if err != nil {
			return token, err
		}
	}
}

type Path struct {
	ID string `xml:"id,attr"`
	D  string `xml:"d,attr"`
}

func (p *Path) String() string {
	return fmt.Sprintf("path: id=%q, d=%#v", p.ID, p.D)
}

func (p *Path) parse(dec *xml.Decoder, token xml.Token) (xml.Token, error) {
	var err error

	curTag := ""
	for {
		switch t := token.(type) {
		case xml.StartElement:
			if curTag != "" || t.Name.Local != "path" {
				return token, fmt.Errorf("Invalid format")
			}
			curTag = t.Name.Local
			for _, a := range t.Attr {
				switch a.Name.Local {
				case "id":
					p.ID = a.Value
				case "d":
					p.D = a.Value
				}
			}
		case xml.EndElement:
			if curTag == "" {
				return token, fmt.Errorf("Invalid format")
			}
			if t.Name.Local != curTag {
				return token, fmt.Errorf("Unknown tag: %q", t.Name.Local)
			}
			curTag = ""
			return dec.Token()
		case xml.ProcInst:
			// nothing to do
		case xml.CharData:
			// nothing to do
		default:
			return token, fmt.Errorf("Error token: %#v", token)
		}

		token, err = dec.Token()
		if err != nil {
			return token, err
		}
	}
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
	ViewBox Box    `xml:"viewBox,attr"`
	Version string `xml:"version,attr"`

	Groups []*Group `xml:"g"`
}

func (s *SvgRoot) String() string {
	return fmt.Sprintf("svg: id=%q, viewBox=%#v, version=%q, groups=%d", s.ID, s.ViewBox, s.Version, len(s.Groups))
}

func (s *SvgRoot) parseGroups(dec *xml.Decoder, token xml.Token) (xml.Token, error) {
	var err error

	for {
		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local != "g" {
				return token, err
			}
			g := new(Group)
			token, err = g.parse(dec, token)
			if err != nil {
				return token, err
			}
			s.Groups = append(s.Groups, g)
		default:
			return token, err
		}
	}
}

func (s *SvgRoot) Parse(dec *xml.Decoder) error {
	var err error
	var token xml.Token

	token, err = dec.Token()
	if err != nil {
		return err
	}

	curTag := ""
	for {
		switch t := token.(type) {
		case xml.StartElement:
			if curTag == "svg" && t.Name.Local == "g" {
				token, err = s.parseGroups(dec, t)
				if err != nil {
					return fmt.Errorf("Invalid groups: %s, %#v", err, token)
				}
				continue
			}
			if curTag != "" || t.Name.Local != "svg" {
				return fmt.Errorf("Invalid format: %#v", token)
			}
			s.XMLName = t.Name
			curTag = t.Name.Local
			for _, a := range t.Attr {
				switch a.Name.Local {
				case "id":
					s.ID = a.Value
				case "viewBox":
					fields := strings.Fields(a.Value)
					if len(fields) != 4 {
						return fmt.Errorf("Invalid box: %#v", token)
					}
					s.ViewBox.X, _ = strconv.Atoi(fields[0])
					s.ViewBox.Y, _ = strconv.Atoi(fields[1])
					s.ViewBox.Width, _ = strconv.Atoi(fields[2])
					s.ViewBox.Height, _ = strconv.Atoi(fields[3])
				case "version":
					s.Version = a.Value
				}
			}
		case xml.EndElement:
			if curTag == "" {
				return fmt.Errorf("Invalid format: %#v", token)
			}
			if t.Name.Local != curTag {
				return fmt.Errorf("Unknown tag: %q", t.Name.Local)
			}
			curTag = ""
			return err
		case xml.ProcInst:
			// nothing to do
		case xml.CharData:
			// nothing to do
		default:
			return fmt.Errorf("Error token: %#v", token)
		}

		token, err = dec.Token()
		if err != nil {
			return err
		}
	}
}

func printIdent(level int) {
	for ; level > 0; level-- {
		fmt.Print(" ")
	}
}

func main() {
	dec := xml.NewDecoder(os.Stdin)

	svg := new(SvgRoot)
	err := svg.Parse(dec)
	if err != nil && err != io.EOF {
		panic(err)
	}

	fmt.Println(svg)
	for _, g := range svg.Groups {
		printIdent(1)
		fmt.Println(g)
		for _, gg := range g.Groups {
			printIdent(2)
			fmt.Println(gg)
			for _, p := range gg.Paths {
				printIdent(3)
				fmt.Println(p)
			}
		}
	}
	fmt.Println("done")
}
