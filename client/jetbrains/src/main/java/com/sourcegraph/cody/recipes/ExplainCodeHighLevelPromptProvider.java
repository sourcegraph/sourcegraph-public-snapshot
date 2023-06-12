package com.sourcegraph.cody.recipes;

public class ExplainCodeHighLevelPromptProvider implements PromptProvider {
  @Override
  public PromptContext getPromptContext(
      String languageName, String selectedText, String truncatedSelectedText) {
    String promptMessage =
        String.format(
            "Explain the following %s code at a high level. Only include details that are essential to an overall understanding of what's happening in the code.\n"
                + PromptMessages.CODE_SNIPPET_IN_LANGUAGE_FORMAT
                + "\n%s",
            languageName,
            languageName.toLowerCase(),
            truncatedSelectedText,
            PromptMessages.MARKDOWN_FORMAT_PROMPT);

    String displayText =
        String.format(
            "Explain the following code at a high level:\n"
                + PromptMessages.CODE_SNIPPET_IN_LANGUAGE_FORMAT,
            languageName,
            selectedText);

    return new PromptContext(promptMessage, displayText);
  }
}
