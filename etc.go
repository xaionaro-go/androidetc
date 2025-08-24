package androidetc

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var defaultSearchDirs = []string{
	"/vendor/etc",
	"/odm/etc",
	"/system/etc",
	"/product/etc",
	"/apex/com.android.media.swcodec/etc",
	"/system/apex/com.android.media.swcodec/etc",
}

func RegisterParser(name string, parserFunc func([]byte) (AbstractDescriptor, error)) {
	parser[name] = parserFunc
}

var parser = map[string]func([]byte) (AbstractDescriptor, error){
	"MediaCodecs": func(b []byte) (AbstractDescriptor, error) {
		return ParseAs[MediaCodecsDescriptor](b)
	},
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func absPath(fileName string, searchDirs []string) string {
	if fileExists(fileName) {
		return fileName
	}

	for _, dir := range searchDirs {
		tryPath := filepath.Join(dir, fileName)
		if fileExists(tryPath) {
			return tryPath
		}
	}
	return ""
}

func ParseFile(
	fileName string,
	searchDirs []string,
) (AbstractDescriptor, error) {
	if searchDirs == nil {
		searchDirs = defaultSearchDirs
	}

	filePath := absPath(fileName, searchDirs)
	if filePath == "" {
		return nil, fmt.Errorf("file %s not found in search dirs: %s",
			fileName, strings.Join(searchDirs, ", "))
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filePath, err)
	}

	var rootName struct{ XMLName xml.Name }
	if err := xml.Unmarshal(data, &rootName); err != nil {
		return nil, fmt.Errorf("unmarshaling root name: %w", err)
	}

	parser := parser[rootName.XMLName.Local]
	if parser != nil {
		return parser(data)
	}

	return nil, fmt.Errorf("unexpected root <%s>", rootName.XMLName.Local)
}

func ParseFileAs[T any](fileName string) (*T, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", fileName, err)
	}
	return ParseAs[T](data)
}

func ParseAs[T any](data []byte) (*T, error) {
	var result T
	if err := xml.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func ParseFileRecursively(
	fileName string,
	searchDirs []string,
) ([]AbstractDescriptor, error) {
	return parseFileRecursively(fileName, searchDirs, map[string]struct{}{})
}

type AbstractDescriptor interface {
	GetXMLName() xml.Name
	GetIncludes() []Include
}

type DescriptorCommons struct {
	XMLName  xml.Name  `xml:"MediaCodecs"`
	Includes []Include `xml:"Include"`
}

func (d *DescriptorCommons) GetXMLName() xml.Name {
	return d.XMLName
}

func (d *DescriptorCommons) GetIncludes() []Include {
	return d.Includes
}

func parseFileRecursively(
	path string,
	searchDirs []string,
	isVisited map[string]struct{},
) ([]AbstractDescriptor, error) {
	if _, ok := isVisited[path]; ok {
		return nil, nil
	}
	isVisited[path] = struct{}{}

	d, err := ParseFile(path, searchDirs)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	var results []AbstractDescriptor
	results = append(results, d)
	for _, inc := range d.GetIncludes() {
		subs, err := ParseFileRecursively(inc.Href, searchDirs)
		if err != nil {
			return nil, fmt.Errorf("parsing included %s: %w", inc.Href, err)
		}
		results = append(results, subs...)
	}
	return results, nil
}
