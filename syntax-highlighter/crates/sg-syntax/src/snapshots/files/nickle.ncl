let range
  | doc "Generate an array of integers in the range [`start`, `end`)."
  | Number -> Number -> Array Number
  = fun start end =>
    if end <= start then
      []
    else
      array.generate (fun x => x + start) (end - start)
in

let is_prime
  | doc "Returns true if the argument is a prime number."
  = fun x => x > 1 && array.all (fun d => x % d != 0) (range 2 (x - 1))
in

let Prime = contract.from_predicate is_prime in

let primes
  | doc "Generate `max` primes using Sieve of Eratosthenes."
  | Number -> Array Prime
  = fun max =>
    let limit = number.pow max (1 / 2) in # sqrt(max)
    let drop_multiples = fun x xs =>
      let to_drop = max
        |> array.generate (fun y => (y + 2) * x)
        |> array.filter (fun y => y <= max) in
        |> array.flatten
      array.filter (fun y => array.all ((!=) y) to_drop) xs in
    let rec loop = fun x xs =>
      if x > limit then
        xs
      else
        loop (x + 1) (drop_multiples x xs) in
    loop 2 (range 2 max) in

{
  run = primes
}
