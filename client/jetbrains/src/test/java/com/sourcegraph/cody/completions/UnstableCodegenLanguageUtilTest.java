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
    String languageId =
        UnstableCodegenLanguageUtil.getModelLanguageId(textDocument(null, "foo.js"));
    assertEquals("javascript", languageId);
  }

  @Test
  public void intellijLanguageIdTakesPriorityIfExtensionUknown() {
    String languageId =
        UnstableCodegenLanguageUtil.getModelLanguageId(textDocument("JAVA", "foo.unknown"));
    assertEquals("java", languageId);
  }

  @Test
  public void intellijLanguageIdTakesPriorityIfSupported() {
    String languageId =
        UnstableCodegenLanguageUtil.getModelLanguageId(textDocument("JAVA", "foo.js"));
    assertEquals("java", languageId);
  }

  @Test
  public void extensionLanguageIdTakesPriorityIfIntelliJUnsupported() {
    String languageId =
        UnstableCodegenLanguageUtil.getModelLanguageId(textDocument("something", "foo.js"));
    assertEquals("javascript", languageId);
  }

  @Test
  public void unsupportedExtensionUsedIfThereAreNoAlternatives() {
    String languageId =
        UnstableCodegenLanguageUtil.getModelLanguageId(textDocument(null, "foo.unknown"));
    assertEquals("unknown", languageId);
  }

  @Test
  public void fallbackReturnedWhenExtensionAndLanguageIdCantBeDetermined() {
    String languageId = UnstableCodegenLanguageUtil.getModelLanguageId(textDocument(null, "foo"));
    assertEquals("no-known-extension-detected", languageId);
  }
}
