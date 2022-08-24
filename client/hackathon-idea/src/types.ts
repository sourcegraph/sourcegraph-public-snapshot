export type Package = string

// Rules for permutation:
// - The package names are sorted alphanumerically, so this is valid:
//     [a][b]
//   But this is not:
//     [b][a]
// - The package names are not allowed to be duplicated.
export type Permutations = Map<Package, Set<Package>>
