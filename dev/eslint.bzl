load("@npm//:eslint/package_json.bzl", eslint_bin = "bin")

def eslint_test(name, deps, data = [], **kwargs):
   eslint_bin.eslint_test(
       name = name,
       args = [
            "$(location {})".format(src)
            for src in data
       ],
       data = deps + data,
       **kwargs,
   )
