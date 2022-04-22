local path_patterns = require("sg.path_patterns")

local exclude_paths = path_patterns.combine {
    path_patterns.segment("example"),
    path_patterns.segment("examples"),
    path_patterns.segment("integration"),
    path_patterns.segment("test"),
    path_patterns.segment("testdata"),
    path_patterns.segment("tests"),
}

return {
    exclude_paths = exclude_paths,
}
