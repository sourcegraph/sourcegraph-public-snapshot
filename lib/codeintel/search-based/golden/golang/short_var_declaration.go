  package golang
  
  import (
   "fmt"
  )
  
  func shortVarDeclaration() {
   message := "Hello!"
// ^^^^^^^ local0-message definition
   message = "Hello!"
// ^^^^^^^ local0-message reference
   message1, message2 := "a", "b"
// ^^^^^^^^ local1-message1 definition
//           ^^^^^^^^ local2-message2 definition
   message1, message2 = message, message
// ^^^^^^^^ local1-message1 reference
//           ^^^^^^^^ local2-message2 reference
//                      ^^^^^^^ local0-message reference
//                               ^^^^^^^ local0-message reference
   fmt.Println(message, message1, message2)
//             ^^^^^^^ local0-message reference
//                      ^^^^^^^^ local1-message1 reference
//                                ^^^^^^^^ local2-message2 reference
  }
  
