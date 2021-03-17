module github.com/paketo-buildpacks/dep-server/actions/get-upstream-dependency/entrypoint

go 1.16

require (
	github.com/paketo-buildpacks/dep-server v0.0.0-00010101000000-000000000000
	github.com/sclevine/spec v1.4.0
	github.com/stretchr/testify v1.7.0
)

replace github.com/paketo-buildpacks/dep-server => ../../../
