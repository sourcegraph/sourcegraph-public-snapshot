local patterns = require("sg.patterns")

local exclude_paths = patterns.path_combine {
    patterns.path_segment("example"),
    patterns.path_segment("examples"),
    patterns.path_segment("integration"),
    patterns.path_segment("test"),
    patterns.path_segment("testdata"),
    patterns.path_segment("tests"),
}

return {
    exclude_paths = exclude_paths,
}
