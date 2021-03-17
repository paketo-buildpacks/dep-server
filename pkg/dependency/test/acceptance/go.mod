module github.com/paketo-buildpacks/dep-server/pkg/dependency/test/acceptance

go 1.16

require (
	github.com/Masterminds/semver v1.5.0
	github.com/paketo-buildpacks/dep-server v0.0.0
	github.com/sclevine/spec v1.4.0
	github.com/stretchr/testify v1.7.0
)

replace github.com/paketo-buildpacks/dep-server => ../../../../
