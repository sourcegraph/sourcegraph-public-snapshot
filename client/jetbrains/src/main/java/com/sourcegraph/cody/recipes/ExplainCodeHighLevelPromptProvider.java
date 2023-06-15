package com.sourcegraph.cody.recipes;

public class ExplainCodeHighLevelPromptProvider implements PromptProvider {
  @Override
  public PromptContext getPromptContext(
      Language language, SelectedText selectedText, TruncatedText truncatedSelectedText) {
    String promptMessage =
        new MessageBuilder(language)
            .appendText(
                "Please explain the following "
                    + language.getValue()
                    + " code at a high level. Only include details that are essential to an overall understanding of what's happening in the code.")
            .appendCodeSnippet(truncatedSelectedText)
            .appendText(PromptMessages.MARKDOWN_FORMAT_PROMPT)
            .build();

    String displayText =
        new MessageBuilder(language)
            .appendText("Explain the following code at a high level:")
            .appendCodeSnippet(selectedText)
            .build();

    return new PromptContext(promptMessage, displayText);
  }
}
