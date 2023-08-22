import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

import com.sourcegraph.cody.autocomplete.AutocompleteDocumentContext;
import org.junit.jupiter.api.Test;

public class AutocompleteDocumentContextTest {
  @Test
  public void skipCompletionIfLineSuffixContainsWordChars() {
    AutocompleteDocumentContext context1 = new AutocompleteDocumentContext("", "foo");
    assertFalse(context1.isCompletionTriggerValid());
    AutocompleteDocumentContext context2 = new AutocompleteDocumentContext("bar", "foo");
    assertFalse(context2.isCompletionTriggerValid());
    AutocompleteDocumentContext context3 = new AutocompleteDocumentContext("bar", " = 123; }");
    assertFalse(context3.isCompletionTriggerValid());
  }

  @Test
  public void skipCompletionIfLinePrefixContainsText() {
    AutocompleteDocumentContext context1 = new AutocompleteDocumentContext("foo", "");
    assertFalse(context1.isCompletionTriggerValid());
    AutocompleteDocumentContext context2 = new AutocompleteDocumentContext("foo", ");");
    assertFalse(context2.isCompletionTriggerValid());
  }

  @Test
  public void skipCompletionIfLinePrefixContainsTextPrecededByWhitespace() {
    AutocompleteDocumentContext context1 = new AutocompleteDocumentContext("  foo", "");
    assertFalse(context1.isCompletionTriggerValid());
    AutocompleteDocumentContext context2 = new AutocompleteDocumentContext("\t\tfoo", ");");
    assertFalse(context2.isCompletionTriggerValid());
  }

  @Test
  public void shouldTriggerCompletionIfLineSuffixIsSpecialCharsOnly() {
    AutocompleteDocumentContext context = new AutocompleteDocumentContext("if(", ") {");
    assertTrue(context.isCompletionTriggerValid());
  }
}
