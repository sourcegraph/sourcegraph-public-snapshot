# Python: code intelligence configuration

This page describes additional configuration that may be needed for Python code intelligence on certain codebases. To enable Python code intelligence, see the [installation documentation](install/index.md).

---

## Configuring pip

The Python language server uses `pip` to fetch dependencies. To configure the behavior of `pip` use the `initializationOptions.pipArgs` field in the language server configuration. This specifies a list of arguments to add to the invocation of `pip`.

For example, if you use both a private package index at `https://python.example.com` and the official Python Package Index, you can set the following in your Sourcegraph configuration:

```json
{
  // ...
  "langservers": [
    {
      "language": "python",
      "initializationOptions": {
        "pipArgs": ["--index-url=https://python.example.com", "--extra-index-url=https://pypi.python.org/simple"]
      }
    }
  ]
  // ...
}
```

---

## Inference of package names

The language server will not run `setup.py` or `pip install`. When it encounters an import, it tries to infer the package name and run `pip download`. (This also avoids running the downloaded package's `setup.py`.) This is expected to work as long as the name of the package on PyPI (or your private package index) is the same as the name that's imported in the source code.
