  package example;
  
  public class spread_parameter {
     String method(String ...a) {
//                           ^ local0-a definition
      return a[0];
//           ^ local0-a reference
    }
  }
  
