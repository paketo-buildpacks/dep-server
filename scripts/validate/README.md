## Validation for new dependendency-management code

A tool for maintainers to validate that the implementation of
dependency-specific code adheres to the expected API and runs as expected.
Validation entails checking that each of the requirements outlined in the
[Context section](#context) takes in the right inputs, and runs reasonably. The
business logic within the dependency code should still be independently
reviewed, this validation solely serves as a way to ensure that the code will work
well with workflows that automate the execution of these steps.

### Quick Start
Run the following if the dependency code you are validating has retrieval, compilation, and test code:
```
cd dep-server/scripts/validate
```
```
go build .
```
```
./validate  --buildpack-dir=<absolute path to buildpack directory of interest> --version <a real dependency version>
```

--------------
Run the following if the dependency code you are validating has retrieval, and test code, but no compilation:
```
cd dep-server/scripts/validate
```
```
go build .
```
Download a real dependency tarball of one of the dependencies, for example `https://deps.paketo.io/bundler/bundler_2.3.22_linux_noarch_bionic_cbc89264.tgz`
```
./validate  --buildpack-dir=<absolute path to buildpack directory of interest> --version <real dependency version> --artifact-path <absolute path to downloaded dependency>
```


### Detailed Usage
1. `go build` the validation code
```
go build .
```

2. Run `./validate --help`

```
Usage of ./validate:
  -artifact-path make test
        OPTIONAL, absolute path to a local artifact to run make test against (if applicable).
         If not provided and compilation runs, compiled tarball will be used for testing.
  -buildpack-dir string
        REQUIRED, Absolute path to buildpack directory
  -version string
        OPTIONAL, version to compile and/or test (if applicable) (default "1.2.3")
```

3. Run validation with arguments

This will validate that `retrieve`, `test` (if there  is a test), and
`compile` (if there is a compilation action) exist, and work correctly.

### `--buildpack-dir`
The `--buildpack-dir` field is REQUIRED, and it should be the absolute path to
a buildpack directory that you want to validate the dependency code for.
Running with the `--buildpack-dir` argument **only** is recommended if there is no
compilation or test code to validate.

Example:
```
./validate  --buildpack-dir="/home/workspace/paketo-buildpacks/bundler"
```

### `--artifact-path`
The `--artifact-path` field is optional, it should be an absolute path to a
tarball to test against. In most cases, the user can download the dependency
and use the path to the dependency for this argument. It's recommended to use
this flag if there is no compilation code, so the `test` code has something to
run against. It should be used with the `--version` flag.

Example:
```
./validate  --buildpack-dir="/home/workspace/paketo-buildpacks/bundler" --artifact-path="~/Downloads/bundler_2.3.22.tgz"
```

### `--version`
The `--version` field is optional, it should be a dependency version, used for
compilation and testing if applicable. The value should be a real version of
the dependency. If passed in with an `--artifact-path` the version should match
the version of the artifact. A default of `1.2.3` will be used if a version is
not provided.

Example:
```
./validate  --buildpack-dir="/home/workspace/paketo-buildpacks/bundler" --artifact-path="~/Downloads/bundler_2.3.22.tgz" --version="2.3.22"
```


### Context
Part of https://github.com/paketo-buildpacks/rfcs/issues/236 entails each
dependency-providing buildpack incorporating new dependency-management related
code that adheres to the API laid out in [Dependency Management RFC Phase
1](https://github.com/paketo-buildpacks/rfcs/blob/main/text/dependencies/rfcs/0004-dependency-management-phase-one.md)

(From the RFC) This includes:
```
buildpack
└───dependency/
│   └───(optional) actions/
│   │   └── compile/
│   │       ├── entrypoint
│   │       ├── action.yml
│   │       ├── <target>.Dockerfile
│   │       └── ...
│   └───retrieval/
│   │   │   ...
│   └───(optional) test/
│   │   │   ...
│   └───Makefile
```
- A `dependency` directory in the buildpack
- `dependency/retrieve` - Code to find new versions of dependencies within `buildpack.toml` constraints and generate metadata for each version, in the form of a `metadata.json`
  - Runs by `make retrieve` in the `dependency/Makefile`
  - (1) Takes in `buildpackTomlPath`, and an `output` name to write metadata to
  - (2) Outputs metadata to the `output` that adheres to schema laid out in RFC
  - (3) If the dependency is to be compiled, the SHA256 and URI are to be left out of metadata.
- Optional `dependency/actions/compile` - Github action to compile dependency
  - Runs by:
  - (1) Build the target-specific Dockerfile
  - (2) Runs compilation via `docker run -v <output dir>:/output compilation --version <version> --outputDir /output --target <target>`
- Optional `dependency/test` - Code to test a dependency artifact
  - Runs by `make test` in the `dependency/Makefile`
  - (1) Takes in a `tarballPath` and `version`
- A `Makefile` to execute `retrieve` and `test`

