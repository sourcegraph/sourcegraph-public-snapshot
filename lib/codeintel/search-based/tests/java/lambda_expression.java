package example;

import java.util.Optional;

public class lambda_expression {
  Optional<Integer> method() {
    Collections.<String, String>emptyMap().compute("1", (a, b) -> a + b);
    return Optional.of(1).map(a -> a + 1);
  }
}
