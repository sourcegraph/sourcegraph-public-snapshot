import static org.junit.jupiter.api.Assertions.*;

import com.sourcegraph.cody.autocomplete.prompt_library.PostProcess;
import org.junit.jupiter.api.Test;

public class PostProcessTest {
  @Test
  public void filtersOddIndentationInSingleLineCompletions() {
    String prefix = "const foo = ";
    String input = " 1";
    String output = PostProcess.postProcess(prefix, input);
    assertEquals("1", output);
  }

  @Test
  public void filtersBadCompletionStart1() {
    String prefix = "one:\n";
    String input = "➕     1";
    String output = PostProcess.postProcess(prefix, input);
    assertEquals("1", output);
  }

  @Test
  public void filtersBadCompletionStart2() {
    String prefix = "one:\n";
    String input = "\u200B   2";
    String output = PostProcess.postProcess(prefix, input);
    assertEquals("2", output);
  }

  @Test
  public void filtersBadCompletionStart3() {
    String prefix = "one:\n";
    String input = ".      3";
    String output = PostProcess.postProcess(prefix, input);
    assertEquals("3", output);
  }

  @Test
  public void filtersBadCompletionStart4() {
    String prefix = "two:\n";
    String input = "+  1";
    String output = PostProcess.postProcess(prefix, input);
    assertEquals("1", output);
  }

  @Test
  public void filtersBadCompletionStart5() {
    String prefix = "two:\n";
    String input = "-  2";
    String output = PostProcess.postProcess(prefix, input);
    assertEquals("2", output);
  }
}
