import static org.junit.jupiter.api.Assertions.*;

import com.sourcegraph.cody.completions.CodyCompletionsManager;
import com.sourcegraph.cody.completions.CompletionDocumentContext;
import com.sourcegraph.cody.vscode.InlineCompletionItem;
import com.sourcegraph.cody.vscode.Position;
import com.sourcegraph.cody.vscode.Range;
import org.junit.jupiter.api.Test;

public class InlineCompletionsPostProcessingTest {

  private final String sameLinePrefix = "System.out.println(\"Hello ";
  private final String sameLineSuffix = "\");";
  private final CompletionDocumentContext testDocumentContext =
      new CompletionDocumentContext(sameLinePrefix, sameLineSuffix);

  @Test
  public void lineSuffixSeparateFromCompletion() {
    String suggestionText = "world!";
    Range inputRange = new Range(new Position(0, 0), new Position(0, 6));
    InlineCompletionItem inputCompletion =
        new InlineCompletionItem(suggestionText, sameLineSuffix, inputRange, null);
    InlineCompletionItem outputCompletion =
        CodyCompletionsManager.postProcessInlineCompletionBasedOnDocumentContext(
            inputCompletion, testDocumentContext);
    assertEquals(outputCompletion.insertText, inputCompletion.insertText);
    assertEquals(outputCompletion.range, inputCompletion.range);
  }

  @Test
  public void completionEndsInLineSuffix() {
    String suggestionTextWithoutSuffix = "world!";
    String suggestionText = suggestionTextWithoutSuffix + sameLineSuffix;
    Range inputRange = new Range(new Position(0, 0), new Position(0, 9));
    InlineCompletionItem inputCompletion =
        new InlineCompletionItem(suggestionText, sameLineSuffix, inputRange, null);
    InlineCompletionItem outputCompletion =
        CodyCompletionsManager.postProcessInlineCompletionBasedOnDocumentContext(
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
    InlineCompletionItem inputCompletion =
        new InlineCompletionItem(suggestionText, sameLineSuffix, inputRange, null);
    InlineCompletionItem outputCompletion =
        CodyCompletionsManager.postProcessInlineCompletionBasedOnDocumentContext(
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
    InlineCompletionItem inputCompletion =
        new InlineCompletionItem(suggestionText, sameLineSuffix, inputRange, null);
    InlineCompletionItem outputCompletion =
        CodyCompletionsManager.removeUndesiredCharacters(inputCompletion);
    assertEquals(outputCompletion.insertText, " world!");
    assertEquals(
        outputCompletion.range,
        inputCompletion.range.withEnd(inputCompletion.range.end.withCharacter(7)));
  }

  @Test
  public void completionContainsLineSeparatorChar() {
    String suggestionText = "\u2028 \u2028world!\u2028";
    Range inputRange = new Range(new Position(0, 0), new Position(0, 10));
    InlineCompletionItem inputCompletion =
        new InlineCompletionItem(suggestionText, sameLineSuffix, inputRange, null);
    InlineCompletionItem outputCompletion =
        CodyCompletionsManager.removeUndesiredCharacters(inputCompletion);
    assertEquals(outputCompletion.insertText, " world!");
    assertEquals(
        outputCompletion.range,
        inputCompletion.range.withEnd(inputCompletion.range.end.withCharacter(7)));
  }
}
