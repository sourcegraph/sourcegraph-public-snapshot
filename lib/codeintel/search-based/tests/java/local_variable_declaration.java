package example;

public class variable_variable_declaration {
  String field = "";
  public final String test() {
    String variable = field;
    String variable2 = variable;
    return variable + variable2;
  }
  class Inner1 {
    public final String test1() {
      String variable = field;
      return variable;
    }
  }
  class Inner2 {
    public final String test2() {
      String variable = field;
      return variable;
    }
  }
  public final String test3() {
    String variable = field;
    return variable;
  }
}
