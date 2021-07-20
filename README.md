# Dep Server

## Summary
This repository is responsible for setting up the Dependencies Server, which
provides a way to access dependencies and their associated metadata. The server
can be accessed at: api.deps.paketo.io.

This repository also contains dependency specific code for retrieving versions
and metadata, which can be found in the
[`pkg/dependency`](https://github.com/paketo-buildpacks/dep-server/tree/main/pkg/dependency)
directory. [Github
Actions](https://github.com/paketo-buildpacks/dep-server/tree/main/actions) and
[workflows](https://github.com/paketo-buildpacks/dep-server/tree/main/.github/workflows)
for automating the retrievial of new dependencies can also be found in this
repository.


## Usage
`curl https://api.deps.paketo.io/v1/dependency?name=<DEP-NAME>` to retrive
dependency metadata.

## Supported Dependencies
* bundler
* composer
* curl
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
* rust
* tini
* yarn

## Example

**Request:**

`curl https://api.deps.paketo.io/v1/dependency?name=go`

**Response:**

```
[
  {
    "name": "go",
    "version": "go1.16.2",
    "sha256": "abd965e71fad990d13d26e737c25a57184a33969e302d723c2b156c84dc619a5",
    "uri": "https://deps.paketo.io/go/go_go1.16.2_linux_x64_bionic_abd965e7.tgz",
    "stacks": [
      {
        "id": "io.buildpacks.stacks.bionic"
      },
      {
        "id": "io.paketo.stacks.tiny"
      }
    ],
    "source": "https://dl.google.com/go/go1.16.2.src.tar.gz",
    "source_sha256": "37ca14287a23cb8ba2ac3f5c3dd8adbc1f7a54b9701a57824bf19a0b271f83ea",
    "deprecation_date": "",
    "created_at": "2021-03-11T20:20:29+00:00",
    "modified_at": "2021-03-11T20:20:29+00:00",
    "cpe": "cpe:2.3:a:golang:go:1.16.2:*:*:*:*:*:*:*",
    "purl": "pkg:generic/go@go1.16.2?download_url=https://dl.google.com/go/go1.16.2.src.tar.gz&checksum=sha256:abd965e71fad990d13d26e737c25a57184a33969e302d723c2b156c84dc619a5",
    "licenses": [
      "BSD-3-Clause"
    ]
  },
  {
    "name": "go",
    "version": "go1.15.10",
    "sha256": "39d03136ebc4d9c230c0a8ca52a9ebbca7d41669f9161a9b5d68cdf7c14a9c40",
    "uri": "https://deps.paketo.io/go/go_go1.15.10_linux_x64_bionic_39d03136.tgz",
    "stacks": [
      {
        "id": "io.buildpacks.stacks.bionic"
      },
      {
        "id": "io.paketo.stacks.tiny"
      }
    ],
    "source": "https://dl.google.com/go/go1.15.10.src.tar.gz",
    "source_sha256": "c1dbca6e0910b41d61a95bf9878f6d6e93d15d884c226b91d9d4b1113c10dd65",
    "deprecation_date": "",
    "created_at": "2021-03-11T20:20:26+00:00",
    "modified_at": "2021-03-11T20:20:26+00:00",
    "cpe": "cpe:2.3:a:golang:go:1.15.10:*:*:*:*:*:*:*",
    "purl": "pkg:generic/go@go1.15.10?download_url=https://dl.google.com/go/go1.15.10.src.tar.gz&checksum=sha256:abd965e71fad990d13d26e737c25a57184a33969e302d723c2b156c84dc619a5",
    "licenses": [
      "BSD-3-Clause"
    ]
  }
]
```
