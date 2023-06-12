package com.sourcegraph.cody.recipes;

public interface PromptProvider {
  PromptContext getPromptContext(
      Language language, SelectedText selectedText, TruncatedText truncatedSelectedText);
}
