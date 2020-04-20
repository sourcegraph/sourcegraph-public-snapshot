# Example: Refactor Go code using Comby

Our goal for this campaign is to simplify Go code by using [Comby](https://comby.dev/) to rewrite calls to `fmt.Sprintf("%d", arg)` with `strconv.Itoa(:[v])`. The semantics are the same, but one more cleanly expresses the intention behind the code.

> **Note**: Learn more about Comby and what it's capable of at [comby.dev](https://comby.dev/)

To do that we use two Docker containers. One container launches Comby to rewrite the the code in Go files and the other runs [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports) to update the `import` statements in the updated Go code so that `strconv` is correctly imported and, possibly, `fmt` is removed.

Here is the `action.json` file that defines this as an action:

```json
{
  "scopeQuery": "lang:go fmt.Sprintf",
  "steps": [
    {
      "type": "docker",
      "image": "comby/comby",
      "args": ["-in-place", "fmt.Sprintf(\"%d\", :[v])", "strconv.Itoa(:[v])", "-matcher", ".go", "-d", "/work"]
    },
    {
      "type": "docker",
      "image": "cytopia/goimports",
      "args": ["-w", "/work"]
    }
  ]
}
```

Please note that the `"scopeQuery"` makes sure that the repositories over which we run the action all contain Go code in which we have a call to `fmt.Sprintf`. That narrows the list of repositories down considerably, even though we still need to search through the whole repository with Comby. (We're aware that this is a limitation and are working on improving the workflows involving exact search results.)

Save the definition in a file, for example `go-comby.action.json`.

Now we can execute the action and turn it into a campaign:

1. Make sure that the `"scopeQuery"` returns the repositories we want to run over: `src actions scope-query -f go-comby.action.json`
1. Execute the action and create a patchset: `src actions exec -f action.json | src campaign patchset create-from-patches`
1. Follow the printed instructions to create and run the campaign on Sourcegraph
