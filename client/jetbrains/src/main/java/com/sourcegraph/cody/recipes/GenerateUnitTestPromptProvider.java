package com.sourcegraph.cody.recipes;

public class GenerateUnitTestPromptProvider implements PromptProvider {
  @Override
  public PromptContext getPromptContext(
      Language language, SelectedText selectedText, TruncatedText truncatedSelectedText) {
    String promptMessage =
        String.format(
            "Generate a unit test in %s for the following code:\n"
                + PromptMessages.CODE_SNIPPET_IN_LANGUAGE_FORMAT
                + "\n%s",
            language.getValue(),
            language.getValue().toLowerCase(),
            truncatedSelectedText.getValue(),
            PromptMessages.MARKDOWN_FORMAT_PROMPT);

    String displayText =
        String.format(
            "Generate a unit test for the following code:\n"
                + PromptMessages.CODE_SNIPPET_IN_LANGUAGE_FORMAT,
            language.getValue(),
            selectedText.getValue());

    return new PromptContext(promptMessage, displayText);
  }
}
