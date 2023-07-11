import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

import com.sourcegraph.cody.autocomplete.AutoCompleteDocumentContext;
import org.junit.jupiter.api.Test;

public class AutoCompleteDocumentContextTest {
  @Test
  public void skipCompletionIfLineSuffixContainsWordChars() {
    AutoCompleteDocumentContext context1 = new AutoCompleteDocumentContext("", "foo");
    assertFalse(context1.isCompletionTriggerValid());
    AutoCompleteDocumentContext context2 = new AutoCompleteDocumentContext("bar", "foo");
    assertFalse(context2.isCompletionTriggerValid());
    AutoCompleteDocumentContext context3 = new AutoCompleteDocumentContext("bar", " = 123; }");
    assertFalse(context3.isCompletionTriggerValid());
  }

  @Test
  public void skipCompletionIfLinePrefixContainsText() {
    AutoCompleteDocumentContext context1 = new AutoCompleteDocumentContext("foo", "");
    assertFalse(context1.isCompletionTriggerValid());
    AutoCompleteDocumentContext context2 = new AutoCompleteDocumentContext("foo", ");");
    assertFalse(context2.isCompletionTriggerValid());
  }

  @Test
  public void skipCompletionIfLinePrefixContainsTextPrecededByWhitespace() {
    AutoCompleteDocumentContext context1 = new AutoCompleteDocumentContext("  foo", "");
    assertFalse(context1.isCompletionTriggerValid());
    AutoCompleteDocumentContext context2 = new AutoCompleteDocumentContext("\t\tfoo", ");");
    assertFalse(context2.isCompletionTriggerValid());
  }

  @Test
  public void shouldTriggerCompletionIfLineSuffixIsSpecialCharsOnly() {
    AutoCompleteDocumentContext context = new AutoCompleteDocumentContext("if(", ") {");
    assertTrue(context.isCompletionTriggerValid());
  }
}
