module github.com/paketo-buildpacks/dep-server/actions/update-php-extension-files/entrypoint

go 1.16

replace github.com/paketo-buildpacks/dep-server => ../../../

require (
	github.com/Masterminds/semver v1.5.0
	github.com/jessevdk/go-flags v1.5.0
	github.com/maxbrunsfeld/counterfeiter/v6 v6.4.1 // indirect
	github.com/onsi/gomega v1.13.0
	github.com/paketo-buildpacks/dep-server v0.0.0-00010101000000-000000000000
	github.com/sclevine/spec v1.4.0
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5 // indirect
	gopkg.in/yaml.v2 v2.4.0
)
