package com.sourcegraph.cody.recipes;

public class ExplainCodeDetailedPromptProvider implements PromptProvider {
  @Override
  public PromptContext getPromptContext(
      Language language, OriginalText originalText, TruncatedText truncatedText) {
    String promptMessage =
        new MessageBuilder(language)
            .appendText(
                "Please explain the following "
                    + language.getValue()
                    + " code. Be very detailed and specific, and indicate when it is not clear to you what is going on. Format your response as an ordered list.")
            .appendCodeSnippet(truncatedText)
            .appendText(PromptMessages.MARKDOWN_FORMAT_PROMPT)
            .build();

    String displayText =
        new MessageBuilder(language)
            .appendText("Explain the following code:")
            .appendCodeSnippet(originalText)
            .build();

    return new PromptContext(promptMessage, displayText);
  }
}
