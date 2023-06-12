package com.sourcegraph.cody.recipes;

public class ExplainCodeDetailedPromptProvider implements PromptProvider {
  @Override
  public PromptContext getPromptContext(
      Language language, SelectedText selectedText, TruncatedText truncatedSelectedText) {
    String promptMessage =
        new MessageBuilder(language)
            .appendText(
                "Please explain the following "
                    + language.getValue()
                    + " code. Be very detailed and specific, and indicate when it is not clear to you what is going on. Format your response as an ordered list.")
            .appendCodeSnippet(truncatedSelectedText)
            .appendText(PromptMessages.MARKDOWN_FORMAT_PROMPT)
            .build();

    String displayText =
        new MessageBuilder(language)
            .appendText("Explain the following code:")
            .appendCodeSnippet(selectedText)
            .build();

    return new PromptContext(promptMessage, displayText);
  }
}
