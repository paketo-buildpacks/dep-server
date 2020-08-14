# Dep Server

## Usage
`curl https://cf-buildpacks-releng.ue.r.appspot.com/api/v1/dependency?name=<DEP-NAME>`

## Example

**Request:**

`curl https://cf-buildpacks-releng.ue.r.appspot.com/api/v1/dependency?name=bundler`

**Response:**

```
[
  {
    "name": "bundler",
    "version": "2.1.4",
    "sha256": "7e7c0a43afe379322b01d62383a760c3d6ed1d633742c3ae5362aa754ffe34c0",
    "uri": "https://pivotal-buildpacks.s3.amazonaws.com/deps/bundler/bundler_2.1.4_linux_noarch_bionic_7e7c0a43.tgz",
    "stacks": [
      {
        "id": "io.buildpacks.stacks.bionic"
      }
    ],
    "source": "https://github.com/bundler/bundler/tree/v2.1.4",
    "source_sha256": "50014d21d6712079da4d6464de12bb93c278f87c9200d0b60ba99f32c25af489",
    "deprecation_date": ""
  },
  {
    "name": "bundler",
    "version": "1.1.2",
    "sha256": "8jasdfja43afe379322b01d62383a760c3d6ed1d633742c3ae5362aa754ffe3h89",
    "uri": "https://pivotal-buildpacks.s3.amazonaws.com/deps/bundler/bundler_1.1.2_linux_noarch_bionic_8jasdfja.tgz",
    "stacks": [
      {
        "id": "io.buildpacks.stacks.bionic"
      }
    ],
    "source": "https://github.com/bundler/bundler/tree/v1.1.2",
    "source_sha256": "20634d21d6712079da4d6464de12bb93c278f87c9200d0b60ba99f32c25au891",
    "deprecation_date": ""
  }
]
```
