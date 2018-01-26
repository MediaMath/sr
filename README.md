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


