import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

import com.sourcegraph.cody.completions.CompletionDocumentContext;
import org.junit.jupiter.api.Test;

public class CompletionDocumentContextTest {
  @Test
  public void skipCompletionIfLineSuffixContainsWordChars() {
    CompletionDocumentContext context1 = new CompletionDocumentContext("", "foo");
    assertFalse(context1.isCompletionTriggerValid());
    CompletionDocumentContext context2 = new CompletionDocumentContext("bar", "foo");
    assertFalse(context2.isCompletionTriggerValid());
    CompletionDocumentContext context3 = new CompletionDocumentContext("bar", " = 123; }");
    assertFalse(context3.isCompletionTriggerValid());
  }

  @Test
  public void skipCompletionIfLinePrefixContainsText() {
    CompletionDocumentContext context1 = new CompletionDocumentContext("foo", "");
    assertFalse(context1.isCompletionTriggerValid());
    CompletionDocumentContext context2 = new CompletionDocumentContext("foo", ");");
    assertFalse(context2.isCompletionTriggerValid());
  }

  @Test
  public void skipCompletionIfLinePrefixContainsTextPrecededByWhitespace() {
    CompletionDocumentContext context1 = new CompletionDocumentContext("  foo", "");
    assertFalse(context1.isCompletionTriggerValid());
    CompletionDocumentContext context2 = new CompletionDocumentContext("\t\tfoo", ");");
    assertFalse(context2.isCompletionTriggerValid());
  }

  @Test
  public void shouldTriggerCompletionIfLineSuffixIsSpecialCharsOnly() {
    CompletionDocumentContext context = new CompletionDocumentContext("if(", ") {");
    assertTrue(context.isCompletionTriggerValid());
  }
}
