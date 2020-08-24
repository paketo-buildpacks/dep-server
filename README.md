# Dep Server

## Usage
`curl https://api.deps.paketo.io/v1/dependency?name=<DEP-NAME>`

## Supported Dependencies
* bundler
* CAAPM
* composer
* dotnet-aspnetcore
* dotnet-runtime
* dotnet-sdk
* go
* httpd
* icu
* nginx
* node
* php
* pip
* pipenv
* python
* ruby
* yarn

## Example

**Request:**

`curl https://api.deps.paketo.io/v1/dependency?name=go`

**Response:**

```
[
  {
    "name": "go",
    "version": "1.15",
    "sha256": "29d4ae84b0cb970442becfe70ee76ce9df67341d15da81b370690fac18111e63",
    "uri": "https://deps.paketo.io/go/go_1.15_linux_x64_bionic_29d4ae84.tgz",
    "stacks": [
      {
        "id": "io.buildpacks.stacks.bionic"
      },
      {
        "id": "io.paketo.stacks.tiny"
      }
    ],
    "source": "https://dl.google.com/go/go1.15.src.tar.gz",
    "source_sha256": "69438f7ed4f532154ffaf878f3dfd83747e7a00b70b3556eddabf7aaee28ac3a",
    "deprecation_date": ""
  },
  {
    "name": "go",
    "version": "1.13.15",
    "sha256": "b4ff131749bea80121374747424f2f02bb7dbdabc69b5aad8cff185f15e1aec9",
    "uri": "https://deps.paketo.io/go/go_1.13.15_linux_x64_bionic_b4ff1317.tgz",
    "stacks": [
      {
        "id": "io.buildpacks.stacks.bionic"
      },
      {
        "id": "io.paketo.stacks.tiny"
      }
    ],
    "source": "https://dl.google.com/go/go1.13.15.src.tar.gz",
    "source_sha256": "5fb43171046cf8784325e67913d55f88a683435071eef8e9da1aa8a1588fcf5d",
    "deprecation_date": ""
  }
]
```
