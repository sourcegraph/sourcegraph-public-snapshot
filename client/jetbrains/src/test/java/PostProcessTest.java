import static org.junit.jupiter.api.Assertions.*;

import com.sourcegraph.cody.autocomplete.prompt_library.PostProcess;
import com.sourcegraph.cody.vscode.Completion;
import org.apache.commons.compress.utils.Lists;
import org.junit.jupiter.api.Test;

public class PostProcessTest {
  @Test
  public void filtersOddIndentationInSingleLineCompletions() {
    String prefix = "const foo = ";
    String suffix = ";";
    Completion input = testCompletion(prefix, " 1");
    Completion output = PostProcess.postProcess(prefix, suffix, input);
    assertEquals("1", output.content);
  }

  @Test
  public void filtersBadCompletionStart1() {
    String prefix = "one:\n";
    String suffix = ";";
    Completion input = testCompletion(prefix, "➕     1");
    Completion output = PostProcess.postProcess(prefix, suffix, input);
    assertEquals("1", output.content);
  }

  @Test
  public void filtersBadCompletionStart2() {
    String prefix = "one:\n";
    String suffix = ";";
    Completion input = testCompletion(prefix, "\u200B   2");
    Completion output = PostProcess.postProcess(prefix, suffix, input);
    assertEquals("2", output.content);
  }

  @Test
  public void filtersBadCompletionStart3() {
    String prefix = "one:\n";
    String suffix = ";";
    Completion input = testCompletion(prefix, ".      3");
    Completion output = PostProcess.postProcess(prefix, suffix, input);
    assertEquals("3", output.content);
  }

  @Test
  public void filtersBadCompletionStart4() {
    String prefix = "two:\n";
    String suffix = ";";
    Completion input = testCompletion(prefix, "+  1");
    Completion output = PostProcess.postProcess(prefix, suffix, input);
    assertEquals("1", output.content);
  }

  @Test
  public void filtersBadCompletionStart5() {
    String prefix = "two:\n";
    String suffix = ";";
    Completion input = testCompletion(prefix, "-  2");
    Completion output = PostProcess.postProcess(prefix, suffix, input);
    assertEquals("2", output.content);
  }

  private Completion testCompletion(String prefix, String content) {
    return new Completion(prefix, Lists.newArrayList(), content, "");
  }
}
