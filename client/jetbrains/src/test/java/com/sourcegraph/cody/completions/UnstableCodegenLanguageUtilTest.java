package com.sourcegraph.cody.completions;

import static org.junit.jupiter.api.Assertions.*;

import com.sourcegraph.cody.vscode.TestTextDocument;
import com.sourcegraph.cody.vscode.TextDocument;
import java.net.URI;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;
import org.junit.jupiter.api.Test;

public class UnstableCodegenLanguageUtilTest {

  private TextDocument textDocument(@Nullable String intelliJLanguageId, @NotNull String fileName) {
    return new TestTextDocument(URI.create("file://" + fileName), fileName, "", intelliJLanguageId);
  }

  @Test
  public void extensionUsedIfIntelliJLanguageIdUndefined() {
    // given
    var input = textDocument(null, "foo.js");

    // when
    String output = UnstableCodegenLanguageUtil.getModelLanguageId(input);

    // then
    assertEquals("javascript", output);
  }

  @Test
  public void intellijLanguageIdTakesPriorityIfExtensionUknown() {
    // given
    var input = textDocument("JAVA", "foo.unknown");

    // when
    String output = UnstableCodegenLanguageUtil.getModelLanguageId(input);

    // then
    assertEquals("java", output);
  }

  @Test
  public void intellijLanguageIdTakesPriorityIfSupported() {
    // given
    var input = textDocument("JAVA", "foo.js");

    // when
    String output = UnstableCodegenLanguageUtil.getModelLanguageId(input);

    // then
    assertEquals("java", output);
  }

  @Test
  public void extensionLanguageIdTakesPriorityIfIntelliJUnsupported() {
    // given
    var input = textDocument("something", "foo.js");

    // when
    String output = UnstableCodegenLanguageUtil.getModelLanguageId(input);

    // then
    assertEquals("javascript", output);
  }

  @Test
  public void unsupportedExtensionUsedIfThereAreNoAlternatives() {
    // given
    var input = textDocument(null, "foo.unknown");

    // when
    String output = UnstableCodegenLanguageUtil.getModelLanguageId(input);

    // then
    assertEquals("unknown", output);
  }

  @Test
  public void fallbackReturnedWhenExtensionAndLanguageIdCantBeDetermined() {
    // given
    var input = textDocument(null, "foo");

    // when
    String output = UnstableCodegenLanguageUtil.getModelLanguageId(input);

    // then
    assertEquals("no-known-extension-detected", output);
  }
}
