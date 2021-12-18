  package golang
  
  import "fmt"
  
  func functionDeclaration(a int, b, c string) {
//                         ^ local0-a definition
//                                ^ local1-b definition
//                                   ^ local2-c definition
   fmt.Println(a, b, c)
//             ^ local0-a reference
//                ^ local1-b reference
//                   ^ local2-c reference
  }
  
