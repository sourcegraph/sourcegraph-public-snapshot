package com.sourcegraph.cody.recipes;

import org.jetbrains.annotations.NotNull;

public class MessageBuilder {

  private final @NotNull Language language;
  private final @NotNull StringBuilder content = new StringBuilder();

  public MessageBuilder(@NotNull Language language) {
    this.language = language;
  }

  public MessageBuilder appendText(String text) {
    content.append(text);
    return this;
  }

  public MessageBuilder appendNewLine() {
    return this.appendText("\n");
  }

  public MessageBuilder appendCodeSnippet(TruncatedText truncatedText) {
    return appendCodeSnippet(truncatedText.getValue());
  }

  public MessageBuilder appendCodeSnippet(OriginalText originalText) {
    return appendCodeSnippet(originalText.getValue());
  }

  private MessageBuilder appendCodeSnippet(String code) {
    return this.appendNewLine()
        .appendText("```")
        .appendText(language.getValue().toLowerCase())
        .appendNewLine()
        .appendText(code)
        .appendNewLine()
        .appendText("```")
        .appendNewLine();
  }

  public String build() {
    return content.toString();
  }
}
