type Package = string

// Rules for permutation:
// - The package names are sorted alphanumerically, so this is valid:
//     [a, b]
//   But this is not:
//     [b, a]
// - The package names are not allowed to be duplicated.
type Permutation = [Package, Package]

type createPermutation = () => Permutation[]

type getPackageInfroamtion = (packageName: Package) => Object
