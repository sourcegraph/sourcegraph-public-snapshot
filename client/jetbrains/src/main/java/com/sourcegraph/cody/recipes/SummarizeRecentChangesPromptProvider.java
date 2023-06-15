package com.sourcegraph.cody.recipes;

public class SummarizeRecentChangesPromptProvider implements PromptProvider {

  private final String displayText;

  public SummarizeRecentChangesPromptProvider(String displayText) {
    this.displayText = displayText;
  }

  @Override
  public PromptContext getPromptContext(
      Language language, OriginalText originalText, TruncatedText truncatedText) {
    String promptMessage =
        new MessageBuilder(language)
            .appendText("Summarize these commits:")
            .appendNewLine()
            .appendText(truncatedText.getValue())
            .appendNewLine()
            .appendText(
                "Provide your response in the form of an unordered list, each commit should be a new list item. Do not mention the commit hashes.")
            .build();
    return new PromptContext(promptMessage, displayText, "Here is a summary of recent changes:\n");
  }
}
