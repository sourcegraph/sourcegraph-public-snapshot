package com.sourcegraph.cody.recipes;

public class GenerateUnitTestPromptProvider implements PromptProvider {
  @Override
  public PromptContext getPromptContext(
      Language language, OriginalText originalText, TruncatedText truncatedText) {
    String promptMessage =
        new MessageBuilder(language)
            .appendText(
                "Generate a unit test in " + language.getValue() + " for the following code:")
            .appendCodeSnippet(truncatedText)
            .appendText(PromptMessages.MARKDOWN_FORMAT_PROMPT)
            .build();

    String displayText =
        new MessageBuilder(language)
            .appendText("Generate a unit test for the following code:")
            .appendCodeSnippet(originalText)
            .build();

    return new PromptContext(promptMessage, displayText);
  }
}
