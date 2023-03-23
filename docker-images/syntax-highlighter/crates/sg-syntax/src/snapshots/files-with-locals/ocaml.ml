module X = struct
  let x = 0
end

let x = 1

let x()=
  let x = x in
  x + X.x
