package svg

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"unicode"
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
	A float32
	B float32
	C float32
	D float32
	E float32
	F float32
}

func (m *Matrix) parse(s string) {
	m.A = 1.0
	m.B = 0
	m.C = 0
	m.D = 1.0
	m.E = 0
	m.F = 0

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
			m.A = float32(a)
			m.B = float32(b)
			m.C = float32(c)
			m.D = float32(d)
			m.E = float32(e)
			m.F = float32(f)
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

type Point struct {
	X float32
	Y float32
}

type Command struct {
	Command string   // m, s, c, l, h/v, z
	Points  []*Point // m[1], s[2], c[3], l[1], h x, v y, z[0]
}

func (c *Command) parse(s string) {
	var err error

	var ps []string
	b := 0
	e := 0
	for {
		ee := strings.IndexAny(s[e:], "-, ")
		if ee == -1 {
			break
		}
		e = e + ee
		if b == e {
			e++
			if s[b] != '-' {
				b = e
			}
			continue
		}
		ps = append(ps, s[b:e])
		b = e
	}
	if b < len(s) {
		ps = append(ps, s[b:])
	}

	switch c.Command {
	case "m", "M", "l", "L":
		if len(ps)&1 != 0 {
			err = fmt.Errorf("Invalid args[%d] of %q command: %s", len(ps), c.Command, s)
			break
		}
		n := len(ps)
		for i := 0; i < n; i += 2 {
			x, _ := strconv.ParseFloat(ps[i+0], 32)
			y, _ := strconv.ParseFloat(ps[i+1], 32)
			p := new(Point)
			p.X = float32(x)
			p.Y = float32(y)
			c.Points = append(c.Points, p)
		}
	case "s", "S":
		if len(ps)&3 != 0 {
			err = fmt.Errorf("Invalid args[%d] of %q command: %s", len(ps), c.Command, s)
			break
		}
		n := len(ps)
		for i := 0; i < n; i += 4 {
			x0, _ := strconv.ParseFloat(ps[i+0], 32)
			y0, _ := strconv.ParseFloat(ps[i+1], 32)
			x1, _ := strconv.ParseFloat(ps[i+2], 32)
			y1, _ := strconv.ParseFloat(ps[i+3], 32)
			p0 := new(Point)
			p0.X = float32(x0)
			p0.Y = float32(y0)
			c.Points = append(c.Points, p0)
			p1 := new(Point)
			p1.X = float32(x1)
			p1.Y = float32(y1)
			c.Points = append(c.Points, p1)
		}
	case "c", "C":
		if len(ps)%6 != 0 {
			err = fmt.Errorf("Invalid args[%d] of %q command: %s", len(ps), c.Command, s)
			break
		}
		n := len(ps)
		for i := 0; i < n; i += 6 {
			x0, _ := strconv.ParseFloat(ps[i+0], 32)
			y0, _ := strconv.ParseFloat(ps[i+1], 32)
			x1, _ := strconv.ParseFloat(ps[i+2], 32)
			y1, _ := strconv.ParseFloat(ps[i+3], 32)
			x2, _ := strconv.ParseFloat(ps[i+4], 32)
			y2, _ := strconv.ParseFloat(ps[i+5], 32)
			p0 := new(Point)
			p0.X = float32(x0)
			p0.Y = float32(y0)
			c.Points = append(c.Points, p0)
			p1 := new(Point)
			p1.X = float32(x1)
			p1.Y = float32(y1)
			c.Points = append(c.Points, p1)
			p2 := new(Point)
			p2.X = float32(x2)
			p2.Y = float32(y2)
			c.Points = append(c.Points, p2)
		}
	case "h", "H":
		if len(ps) != 1 {
			err = fmt.Errorf("Invalid args[%d] of %q command: %s", len(ps), c.Command, s)
			break
		}
		x, _ := strconv.ParseFloat(ps[0], 32)
		p := new(Point)
		p.X = float32(x)
		c.Points = append(c.Points, p)
	case "v", "V":
		if len(ps) != 1 {
			err = fmt.Errorf("Invalid args[%d] of %q command: %s", len(ps), c.Command, s)
			break
		}
		y, _ := strconv.ParseFloat(ps[0], 32)
		p := new(Point)
		p.Y = float32(y)
		c.Points = append(c.Points, p)
	case "z", "Z":
		// nothing to do
	default:
		err = fmt.Errorf("Invalid path command %q: %s", c.Command, s)
	}

	if err != nil {
		panic(err)
	}
}

func (c *Command) String() string {
	return fmt.Sprintf("command: %s, points=%d", c.Command, len(c.Points))
}

type Path struct {
	ID string     `xml:"id,attr"`
	D  []*Command `xml:"d,attr"`
}

func (p *Path) String() string {
	return fmt.Sprintf("path: id=%q, d=%d", p.ID, len(p.D))
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
					fs := strings.FieldsFunc(a.Value, unicode.IsLetter)
					s := a.Value
					for _, field := range fs {
						cmd := new(Command)
						cmd.Command = s[0:1]
						cmd.parse(field)
						p.D = append(p.D, cmd)

						s = s[1+len(field):]
					}
					if len(s) > 0 {
						cmd := new(Command)
						cmd.Command = s
						p.D = append(p.D, cmd)
					}
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

type Root struct {
	XMLName xml.Name `xml:"svg"`

	ID      string `xml:"id,attr"`
	ViewBox Box    `xml:"viewBox,attr"`
	Version string `xml:"version,attr"`

	Groups []*Group `xml:"g"`
}

func (r *Root) String() string {
	return fmt.Sprintf("svg: id=%q, viewBox=%#v, version=%q, groups=%d", r.ID, r.ViewBox, r.Version, len(r.Groups))
}

func (r *Root) parseGroups(dec *xml.Decoder, token xml.Token) (xml.Token, error) {
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
			r.Groups = append(r.Groups, g)
		default:
			return token, err
		}
	}
}

func (r *Root) Parse(dec *xml.Decoder) error {
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
				token, err = r.parseGroups(dec, t)
				if err != nil {
					return fmt.Errorf("Invalid groups: %s, %#v", err, token)
				}
				continue
			}
			if curTag != "" || t.Name.Local != "svg" {
				return fmt.Errorf("Invalid format: %#v", token)
			}
			r.XMLName = t.Name
			curTag = t.Name.Local
			for _, a := range t.Attr {
				switch a.Name.Local {
				case "id":
					r.ID = a.Value
				case "viewBox":
					fields := strings.Fields(a.Value)
					if len(fields) != 4 {
						return fmt.Errorf("Invalid box: %#v", token)
					}
					r.ViewBox.X, _ = strconv.Atoi(fields[0])
					r.ViewBox.Y, _ = strconv.Atoi(fields[1])
					r.ViewBox.Width, _ = strconv.Atoi(fields[2])
					r.ViewBox.Height, _ = strconv.Atoi(fields[3])
				case "version":
					r.Version = a.Value
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
