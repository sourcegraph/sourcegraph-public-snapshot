  package example;
  
  public class variable_variable_declaration {
    String field = "";
    public final String test() {
      String variable = field;
//           ^^^^^^^^ local0-variable definition
      String variable2 = variable;
//           ^^^^^^^^^ local1-variable2 definition
//                       ^^^^^^^^ local0-variable reference
      return variable + variable2;
//           ^^^^^^^^ local0-variable reference
//                      ^^^^^^^^^ local1-variable2 reference
    }
    class Inner1 {
      public final String test1() {
        String variable = field;
//             ^^^^^^^^ local2-variable definition
        return variable;
//             ^^^^^^^^ local2-variable reference
      }
    }
    class Inner2 {
      public final String test2() {
        String variable = field;
//             ^^^^^^^^ local3-variable definition
        return variable;
//             ^^^^^^^^ local3-variable reference
      }
    }
    public final String test3() {
      String variable = field;
//           ^^^^^^^^ local4-variable definition
      return variable;
//           ^^^^^^^^ local4-variable reference
    }
  }
  
