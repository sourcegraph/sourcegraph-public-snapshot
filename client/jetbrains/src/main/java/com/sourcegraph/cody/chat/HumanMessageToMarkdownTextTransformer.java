package com.sourcegraph.cody.chat;

import org.jetbrains.annotations.NotNull;

public class HumanMessageToMarkdownTextTransformer {
  private @NotNull final String humanMessageText;

  private static final String SPACE = " ";

  public HumanMessageToMarkdownTextTransformer(@NotNull String humanMessageText) {
    this.humanMessageText = humanMessageText;
  }

  public String transform() {
    return humanMessageText.replace("\n", "<br />");
  }
}
