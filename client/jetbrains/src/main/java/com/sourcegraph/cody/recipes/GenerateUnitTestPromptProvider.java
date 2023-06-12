package com.sourcegraph.cody.recipes;

public class GenerateUnitTestPromptProvider implements PromptProvider {
  @Override
  public PromptContext getPromptContext(
      String languageName, String selectedText, String truncatedSelectedText) {
    String promptMessage =
        String.format(
            "Generate a unit test in %s for the following code:\n"
                + PromptMessages.CODE_SNIPPET_IN_LANGUAGE_FORMAT
                + "\n%s",
            languageName,
            languageName.toLowerCase(),
            truncatedSelectedText,
            PromptMessages.MARKDOWN_FORMAT_PROMPT);

    String displayText =
        String.format(
            "Generate a unit test for the following code:\n"
                + PromptMessages.CODE_SNIPPET_IN_LANGUAGE_FORMAT,
            languageName,
            selectedText);

    return new PromptContext(promptMessage, displayText);
  }
}
