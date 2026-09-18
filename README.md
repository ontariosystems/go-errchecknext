# go-errchecknext
A lint tool to check that when an `error` is assigned, the next statement should check the error

## Usage
Use this tool by running:

```shell
go run github.com/ontariosystems/go-errchecknext@latest -o lint-report-errchecknext.xml ./...
```

This can also be used as a hook for [pre-commit](https://pre-commit.com/)

```yaml
repos:
  - repo: https://github.com/ontariosystems/go-errchecknext
    rev: v0.1.0
    hooks:
      - id: go-errchecknext
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

See [testdata](internal/testdata) for more good and bad examples
