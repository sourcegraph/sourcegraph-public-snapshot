local contains = function(table, element)
  for i = 1, #table do
    if table[i] == element then
      return true
    end
  end

  return false
end

local contains_any = function(paths, candidates)
  for _, candidate in ipairs(candidates) do
    if contains(paths, candidate) then
      return true
    end
  end

  return false
end

local reverse = function(slice)
  local reversed = {}
  for i = 1, #slice do
    table.insert(reversed, slice[#slice - i + 1])
  end

  return reversed
end

local with_new_head = function(slice, element)
  local new = { element }
  for _, v in ipairs(slice) do
    table.insert(new, v)
  end

  return new
end

return {
  contains = contains,
  contains_any = contains_any,
  reverse = reverse,
  with_new_head = with_new_head,
}
