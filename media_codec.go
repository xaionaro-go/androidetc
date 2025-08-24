package androidetc

import (
	"encoding/xml"
	"fmt"
	"os"
	"path"
	"strings"
)

type MediaCodecsDescriptor struct {
	DescriptorCommons
	Decoders []MediaCodec `xml:"Decoders>MediaCodec"`
	Encoders []MediaCodec `xml:"Encoders>MediaCodec"`
	Settings []Setting    `xml:"Settings>Setting"`
}

type MediaCodecsDescriptors []*MediaCodecsDescriptor

func (m MediaCodecsDescriptors) Decoders() []MediaCodec {
	var all []MediaCodec
	for _, d := range m {
		all = append(all, d.Decoders...)
	}
	return all
}

func (m MediaCodecsDescriptors) Encoders() []MediaCodec {
	var all []MediaCodec
	for _, d := range m {
		all = append(all, d.Encoders...)
	}
	return all
}

type Include struct {
	Href string `xml:"href,attr"`
}

type Setting struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type MediaCodec struct {
	Name     string      `xml:"name,attr"`
	Domain   string      `xml:"domain,attr,omitempty"`
	Types    []CodecType `xml:"Type"`
	Quirks   []Quirk     `xml:"Quirk"`
	Limits   []Limit     `xml:"Limit"`
	Features []Feature   `xml:"Feature"`
}

type CodecType struct {
	Name     string    `xml:"name,attr"`
	Limits   []Limit   `xml:"Limit"`
	Features []Feature `xml:"Feature"`
}

type Quirk struct {
	Name string `xml:"name,attr"`
}

type Limit struct {
	Name  string     `xml:"name,attr"`
	Attrs []xml.Attr `xml:",any,attr"`
}

func (l Limit) Attr(name string) (string, bool) {
	for _, a := range l.Attrs {
		if a.Name.Local == name {
			return a.Value, true
		}
	}
	return "", false
}

type Feature struct {
	Name   string     `xml:"name,attr"`
	Params []xml.Attr `xml:",any,attr"`
}

var mediaCodecSWPrefixes = []string{"omx.google.", "c2.android."}
var mediaCodecHWPrefixes = []string{"omx.qcom.", "omx.exynos.", "omx.mtk.", "omx.hisi.", "c2.qti.", "c2.exynos.", "c2.mtk.", "c2.hisi."}

// IsHardware is a heuristic to determine if this codec is hardware-accelerated.
func (mc MediaCodec) IsHardware() bool {
	if strings.EqualFold(mc.Domain, "software") {
		return false
	}

	n := strings.ToLower(mc.Name)
	for _, p := range mediaCodecSWPrefixes {
		if strings.HasPrefix(n, p) {
			return false
		}
	}

	for _, p := range mediaCodecHWPrefixes {
		if strings.HasPrefix(n, p) {
			return true
		}
	}

	return false
}

func ParseMediaCodecs() (MediaCodecsDescriptors, error) {
	var files []string
	for _, dir := range defaultSearchDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if strings.HasPrefix(entry.Name(), "media_codecs") && strings.HasSuffix(entry.Name(), ".xml") {
				files = append(files, path.Join(dir, entry.Name()))
			}
		}
	}

	var result MediaCodecsDescriptors
	for _, file := range files {
		desc, err := ParseFileAs[MediaCodecsDescriptor](file)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", file, err)
		}

		result = append(result, desc)
	}

	return result, nil
}
