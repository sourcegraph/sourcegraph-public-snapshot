  package golang
  
  import (
   "sort"
  )
  
  func funcLiteral() {
   sort.SliceStable(
    []string{},
    func(i, j int) bool {
//       ^ local0-i definition
//          ^ local1-j definition
     return i < j
//          ^ local0-i reference
//              ^ local1-j reference
    },
   )
  }
  
