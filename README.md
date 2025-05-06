# Run

try to run the tests with the command below:

```console
go test -count=1 -v ./... 2>&1 | tee example.log 
```

important set `-count=1` to bypass any test cache
