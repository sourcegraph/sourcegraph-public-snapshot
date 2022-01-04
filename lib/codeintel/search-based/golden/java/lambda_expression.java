  package example;
  
  import java.util.Optional;
  
  public class lambda_expression {
    Optional<Integer> method() {
      Collections.<String, String>emptyMap().compute("1", (a, b) -> a + b);
//                                                         ^ local0-a definition
//                                                            ^ local1-b definition
//                                                                  ^ local0-a reference
//                                                                      ^ local1-b reference
      return Optional.of(1).map(a -> a + 1);
//                              ^ local2-a definition
//                                   ^ local2-a reference
    }
  }
  
