module github.com/paketo-buildpacks/dep-server/pkg/dependency/test/acceptance

go 1.15

require (
	github.com/paketo-buildpacks/dep-server v0.0.0
	github.com/sclevine/spec v1.4.0
	github.com/stretchr/testify v1.6.1
)

replace github.com/paketo-buildpacks/dep-server => ../../../../
