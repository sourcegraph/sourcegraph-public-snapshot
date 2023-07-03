package com.sourcegraph.cody.completions;

import static org.junit.jupiter.api.Assertions.*;

import org.junit.jupiter.api.Test;

public class UnstableCodegenLanguageUtilTest {

  @Test
  public void extensionUsedIfIntelliJLanguageIdUndefined() {
    String languageId = UnstableCodegenLanguageUtil.getModelLanguageId(null, "foo.js");
    assertEquals("javascript", languageId);
  }

  @Test
  public void intellijLanguageIdTakesPriorityIfExtensionUknown() {
    String languageId = UnstableCodegenLanguageUtil.getModelLanguageId("JAVA", "foo.unknown");
    assertEquals("java", languageId);
  }

  @Test
  public void intellijLanguageIdTakesPriorityIfSupported() {
    String languageId = UnstableCodegenLanguageUtil.getModelLanguageId("JAVA", "foo.js");
    assertEquals("java", languageId);
  }

  @Test
  public void extensionLanguageIdTakesPriorityIfIntelliJUnsupported() {
    String languageId = UnstableCodegenLanguageUtil.getModelLanguageId("something", "foo.js");
    assertEquals("javascript", languageId);
  }

  @Test
  public void unsupportedExtensionUsedIfThereAreNoAlternatives() {
    String languageId = UnstableCodegenLanguageUtil.getModelLanguageId(null, "foo.unknown");
    assertEquals("unknown", languageId);
  }

  @Test
  public void fallbackReturnedWhenExtensionAndLanguageIdCantBeDetermined() {
    String languageId = UnstableCodegenLanguageUtil.getModelLanguageId(null, "foo");
    assertEquals("no-known-extension-detected", languageId);
  }
}
