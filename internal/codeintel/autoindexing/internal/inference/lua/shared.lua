local pattern = require "sg.autoindex.patterns"

local exclude_paths = pattern.new_path_combine {
  pattern.new_path_segment "example",
  pattern.new_path_segment "examples",
  pattern.new_path_segment "integration",
  pattern.new_path_segment "test",
  pattern.new_path_segment "testdata",
  pattern.new_path_segment "tests",
}

return {
  exclude_paths = exclude_paths,
}
