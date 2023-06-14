package com.sourcegraph.cody.recipes;

public class GenerateUnitTestPromptProvider implements PromptProvider {
  @Override
  public PromptContext getPromptContext(
      Language language, SelectedText selectedText, TruncatedText truncatedSelectedText) {
    String promptMessage =
        new MessageBuilder(language)
            .appendText(
                "Generate a unit test in " + language.getValue() + " for the following code:")
            .appendCodeSnippet(truncatedSelectedText)
            .appendText(PromptMessages.MARKDOWN_FORMAT_PROMPT)
            .build();

    String displayText =
        new MessageBuilder(language)
            .appendText("Generate a unit test for the following code:")
            .appendCodeSnippet(selectedText)
            .build();

    return new PromptContext(promptMessage, displayText);
  }
}
