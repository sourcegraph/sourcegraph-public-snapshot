import static org.junit.jupiter.api.Assertions.*;

import com.intellij.psi.codeStyle.CommonCodeStyleSettings;
import com.sourcegraph.cody.autocomplete.AutoCompleteDocumentContext;
import com.sourcegraph.cody.autocomplete.CodyAutoCompleteManager;
import com.sourcegraph.cody.vscode.InlineAutoCompleteItem;
import com.sourcegraph.cody.vscode.Position;
import com.sourcegraph.cody.vscode.Range;
import org.junit.jupiter.api.Test;

public class InlineCompletionsPostProcessingTest {

  private final String sameLinePrefix = "System.out.println(\"Hello ";
  private final String sameLineSuffix = "\");";
  private final AutoCompleteDocumentContext testDocumentContext =
      new AutoCompleteDocumentContext(sameLinePrefix, sameLineSuffix);

  @Test
  public void lineSuffixSeparateFromCompletion() {
    String suggestionText = "world!";
    Range inputRange = new Range(new Position(0, 0), new Position(0, 6));
    InlineAutoCompleteItem inputCompletion =
        new InlineAutoCompleteItem(suggestionText, sameLineSuffix, inputRange, null);
    InlineAutoCompleteItem outputCompletion =
        CodyAutoCompleteManager.postProcessInlineAutoCompleteBasedOnDocumentContext(
            inputCompletion, testDocumentContext);
    assertEquals(outputCompletion.insertText, inputCompletion.insertText);
    assertEquals(outputCompletion.range, inputCompletion.range);
  }

  @Test
  public void completionEndsInLineSuffix() {
    String suggestionTextWithoutSuffix = "world!";
    String suggestionText = suggestionTextWithoutSuffix + sameLineSuffix;
    Range inputRange = new Range(new Position(0, 0), new Position(0, 9));
    InlineAutoCompleteItem inputCompletion =
        new InlineAutoCompleteItem(suggestionText, sameLineSuffix, inputRange, null);
    InlineAutoCompleteItem outputCompletion =
        CodyAutoCompleteManager.postProcessInlineAutoCompleteBasedOnDocumentContext(
            inputCompletion, testDocumentContext);
    assertEquals(outputCompletion.insertText, suggestionTextWithoutSuffix);
    assertEquals(
        outputCompletion.range,
        inputCompletion.range.withEnd(inputCompletion.range.end.withCharacter(6)));
  }

  @Test
  public void completionContainsLineSuffix() {
    String suggestionTextWithoutSuffix = "world!";
    String suggestionText = suggestionTextWithoutSuffix + sameLineSuffix + " // prints hello world";
    Range inputRange = new Range(new Position(0, 0), new Position(0, 31));
    InlineAutoCompleteItem inputCompletion =
        new InlineAutoCompleteItem(suggestionText, sameLineSuffix, inputRange, null);
    InlineAutoCompleteItem outputCompletion =
        CodyAutoCompleteManager.postProcessInlineAutoCompleteBasedOnDocumentContext(
            inputCompletion, testDocumentContext);
    assertEquals(outputCompletion.insertText, suggestionTextWithoutSuffix);
    assertEquals(
        outputCompletion.range,
        inputCompletion.range.withEnd(inputCompletion.range.end.withCharacter(6)));
  }

  @Test
  public void completionContainsZeroWidthSpaces() {
    String suggestionText = "\u200b \u200bworld!\u200b";
    Range inputRange = new Range(new Position(0, 0), new Position(0, 10));
    InlineAutoCompleteItem inputCompletion =
        new InlineAutoCompleteItem(suggestionText, sameLineSuffix, inputRange, null);
    InlineAutoCompleteItem outputCompletion =
        CodyAutoCompleteManager.removeUndesiredCharacters(inputCompletion);
    assertEquals(outputCompletion.insertText, " world!");
    assertEquals(
        outputCompletion.range,
        inputCompletion.range.withEnd(inputCompletion.range.end.withCharacter(7)));
  }

  @Test
  public void completionContainsLineSeparatorChar() {
    String suggestionText = "\u2028 \u2028world!\u2028";
    Range inputRange = new Range(new Position(0, 0), new Position(0, 10));
    InlineAutoCompleteItem inputCompletion =
        new InlineAutoCompleteItem(suggestionText, sameLineSuffix, inputRange, null);
    InlineAutoCompleteItem outputCompletion =
        CodyAutoCompleteManager.removeUndesiredCharacters(inputCompletion);
    assertEquals(outputCompletion.insertText, " world!");
    assertEquals(
        outputCompletion.range,
        inputCompletion.range.withEnd(inputCompletion.range.end.withCharacter(7)));
  }

  @Test
  public void convertCompletionIndentationTabsToSpaces() {
    String suggestionText = "\t    \tHello world! \tHello once again!";
    Range inputRange = new Range(new Position(0, 0), new Position(0, 37));
    InlineAutoCompleteItem inputCompletion =
        new InlineAutoCompleteItem(suggestionText, sameLineSuffix, inputRange, null);
    CommonCodeStyleSettings.IndentOptions indentOptions = // default indent options use tabSize = 4
        CommonCodeStyleSettings.IndentOptions.DEFAULT_INDENT_OPTIONS;
    InlineAutoCompleteItem outputCompletion =
        CodyAutoCompleteManager.normalizeIndentation(inputCompletion, indentOptions);
    String expectedSuggestionText = "            Hello world! \tHello once again!";
    Range expectedRange = inputRange.withEnd(inputRange.end.withCharacter(43));
    assertEquals(outputCompletion.insertText, expectedSuggestionText);
    assertEquals(outputCompletion.range, expectedRange);
  }
}
