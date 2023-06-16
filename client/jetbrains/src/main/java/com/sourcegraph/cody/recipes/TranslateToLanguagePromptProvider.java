package com.sourcegraph.cody.recipes;

public class TranslateToLanguagePromptProvider implements PromptProvider {

  private final Language toLanguage;

  public TranslateToLanguagePromptProvider(Language toLanguage) {
    this.toLanguage = toLanguage;
  }

  @Override
  public PromptContext getPromptContext(
      Language language, OriginalText originalText, TruncatedText truncatedText) {
    String promptMessage =
        new MessageBuilder(language)
            .appendText("Translate the following code into " + toLanguage.getValue())
            .appendCodeSnippet(truncatedText)
            .appendText(PromptMessages.MARKDOWN_FORMAT_PROMPT)
            .build();
    String displayText =
        new MessageBuilder(language)
            .appendText("Translate the following code into " + toLanguage.getValue())
            .appendCodeSnippet(originalText)
            .build();
    return new PromptContext(promptMessage, displayText);
  }
}
