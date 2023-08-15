local recognizers = require "internal_recognizers"

local M = {}

local normalize = function(config)
  local og = config.generate
  if og ~= nil then
    config.generate = function(api, paths, contents_by_path)
      local paths2 = {}
      for i = 1, #paths do
        table.insert(paths2, paths[i])
      end

      return og(api, paths2, contents_by_path)
    end
  end

  local oh = config.hints
  if oh ~= nil then
    config.hints = function(api, paths)
      local paths2 = {}
      for i = 1, #paths do
        table.insert(paths2, paths[i])
      end

      return oh(api, paths2)
    end
  end

  return config
end

-- type: ({
--   "patterns": array[pattern],
--   "patterns_for_content": array[pattern],
--   "generate": (registration_api, paths: array[string], contents_by_path: table[string, string]) -> void,
--   "hints": (registration_api, paths: array[string]) -> void
-- }) -> recognizer
M.new_path_recognizer = function(config)
  return recognizers.path_recognizer(normalize(config))
end

-- type: (array[recognizer]) -> recognizer
M.new_fallback_recognizer = function(config)
  return recognizers.fallback_recognizer(config)
end

return M
