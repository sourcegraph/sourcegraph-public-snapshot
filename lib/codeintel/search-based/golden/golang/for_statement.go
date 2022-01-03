  package golang
  
  import "fmt"
  
  func forStatement() {
  
   for i := 0; i < 10; i++ {
//     ^ local0-i definition
//             ^ local0-i reference
//                     ^ local0-i reference
    fmt.Println(i)
   }
  
   for i, j := 0, 1; i < 10; i, j = i+1, j+2 {
//     ^ local1-i definition
//        ^ local2-j definition
//                   ^ local1-i reference
//                           ^ local1-i reference
//                              ^ local2-j reference
//                                  ^ local1-i reference
//                                       ^ local2-j reference
    fmt.Println(i, j)
   }
  
   for n := range make(chan int, 1) {
//     ^ local3-n definition
    fmt.Println(n)
//              ^ local3-n reference
   }
  
   for i, e := range []string{} {
//     ^ local4-i definition
//        ^ local5-e definition
    fmt.Println(i, e)
//              ^ local4-i reference
//                 ^ local5-e reference
   }
  }
  
