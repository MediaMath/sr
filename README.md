# [sr](https://github.com/MediaMath/sr) &middot; [![CircleCI Status](https://circleci.com/gh/MediaMath/sr.svg?style=shield)](https://circleci.com/gh/MediaMath/sr) [![GitHub license](https://img.shields.io/badge/license-BSD3-blue.svg)](https://github.com/MediaMath/sr/blob/master/LICENSE) [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/MediaMath/sr/blob/master/CONTRIBUTING.md)

# sr - simple library and CLI wrapper around confluent schema registry api

```bash
$ export SCHEMA_REGISTRY_URL=http://example.com
$ sr ls
["foo"]
$ sr add bar ~/Desktop/mt_event.json
{id:"998"}
$ sr ls
["foo", "bar"]
$ sr ls bar
[1]
$ sr ls bar 1
...schema that was added and version and name...
```

```go
func main() {
   id, err := sr.Register(http.DefaultClient, "http://example.com", sr.Subject("foo"), sr.Schema(`{"type":"long"}`))
}
```
