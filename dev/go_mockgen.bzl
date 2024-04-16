load("//dev:go_mockgen_rule.bzl", "go_mockgen_generate")
load("//dev:write_generated_to_source_files.bzl", "write_generated_to_source_files")

def go_mockgen(name, manifests, deps, out):
    gen_file = "_" + out

    go_mockgen_generate(
        name = name + "_generate",
        deps = deps,
        out = gen_file,
        manifests = manifests,
    )

    write_generated_to_source_files(
        name = name,
        output_files = {out: gen_file},
        target = ":" + name + "_generate",
    )
