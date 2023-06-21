package com.sourcegraph.cody.agent.protocol;

import org.jetbrains.annotations.Nullable;

public class TextDocument {

  public String filePath;
  @Nullable public String content;
  @Nullable public Range selection;

  public TextDocument setFilePath(String filePath) {
    this.filePath = filePath;
    return this;
  }

  public TextDocument setContent(String content) {
    this.content = content;
    return this;
  }

  public TextDocument setSelection(Range selection) {
    this.selection = selection;
    return this;
  }

  @Override
  public String toString() {
    return "TextDocument{"
        + "filePath='"
        + filePath
        + '\''
        + ", content='"
        + content
        + '\''
        + ", selection="
        + selection
        + '}';
  }
}
