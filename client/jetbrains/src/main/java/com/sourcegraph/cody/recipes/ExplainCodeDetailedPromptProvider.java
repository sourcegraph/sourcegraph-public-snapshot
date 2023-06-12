package com.sourcegraph.cody.recipes;

public class ExplainCodeDetailedPromptProvider implements PromptProvider {
  @Override
  public PromptContext getPromptContext(
      String languageName, String selectedText, String truncatedSelectedText) {
    String promptMessage =
        String.format(
            "Please explain the following %s code. Be very detailed and specific, and indicate when it is not clear to you what is going on. Format your response as an ordered list.\n"
                + PromptMessages.CODE_SNIPPET_IN_LANGUAGE_FORMAT
                + "\n%s",
            languageName,
            languageName.toLowerCase(),
            truncatedSelectedText,
            PromptMessages.MARKDOWN_FORMAT_PROMPT);

    String displayText =
        String.format(
            "Explain the following code:\n" + PromptMessages.CODE_SNIPPET_IN_LANGUAGE_FORMAT,
            languageName,
            selectedText);

    return new PromptContext(promptMessage, displayText);
  }
}
