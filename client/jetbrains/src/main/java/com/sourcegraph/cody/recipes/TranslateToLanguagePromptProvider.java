package com.sourcegraph.cody.recipes;

public class TranslateToLanguagePromptProvider implements PromptProvider {

  private final Language toLanguage;

  public TranslateToLanguagePromptProvider(Language toLanguage) {
    this.toLanguage = toLanguage;
  }

  @Override
  public PromptContext getPromptContext(
      Language language, SelectedText selectedText, TruncatedText truncatedSelectedText) {
    String promptMessage =
        new MessageBuilder(language)
            .appendText("Translate the following code into " + toLanguage.getValue())
            .appendCodeSnippet(truncatedSelectedText)
            .appendText(PromptMessages.MARKDOWN_FORMAT_PROMPT)
            .build();
    String displayText =
        new MessageBuilder(language)
            .appendText("Translate the following code into " + toLanguage.getValue())
            .appendCodeSnippet(selectedText)
            .build();
    return new PromptContext(promptMessage, displayText);
  }
}
