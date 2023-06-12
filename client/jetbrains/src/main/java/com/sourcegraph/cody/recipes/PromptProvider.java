package com.sourcegraph.cody.recipes;

public interface PromptProvider {
  PromptContext getPromptContext(
      String languageName, String selectedText, String truncatedSelectedText);
}
