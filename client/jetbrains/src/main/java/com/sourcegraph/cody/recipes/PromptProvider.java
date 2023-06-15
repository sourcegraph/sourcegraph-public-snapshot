package com.sourcegraph.cody.recipes;

public interface PromptProvider {
  PromptContext getPromptContext(
      Language language, OriginalText originalText, TruncatedText truncatedText);
}
