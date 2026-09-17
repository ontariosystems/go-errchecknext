# go-errchecknext
A lint tool to check that when an `error` is assigned, the next statement should check the error

## Usage
Use this tool by running:

```shell
go run github.com/ontariosystems/go-errchecknext -o lint-report-errchecknext.xml ./...
```

## Examples
Here's some examples where the linter will find problems.

It is generally bad to do anything other than checking the error after it is assigned
```go
// BAD - do not do this
resp, err := someFunc()
defer resp.Body.Close()
if err != nil {
	return err
}
```

In that example, if `err` is not nil, there is a chance that `resp` is not valid and should not be
used. The error should be processed first.

```go
// GOOD
resp, err := someFunc()
if err != nil {
    return err
}
defer resp.Body.Close()
```

See [testdata](internal/testdata/src) for more good and bad examples