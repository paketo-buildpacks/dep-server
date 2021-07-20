package purl

import (
	"github.com/package-url/packageurl-go"
)

type PURLGenerator struct{}

func NewPURLGenerator() PURLGenerator {
	return PURLGenerator{}
}

func (PURLGenerator) Generate(name, version, sourceSHA, source string) string {
	purl := packageurl.NewPackageURL(
		packageurl.TypeGeneric,
		"",
		name,
		version,
		packageurl.QualifiersFromMap(map[string]string{
			"checksum":     sourceSHA,
			"download_url": source,
		}),
		"",
	)

	return purl.ToString()
}
