package com.sourcegraph.cody.vscode;

public class SelectedAutoCompleteSuggestionInfo {
  public final Range range;
  public final String text;

  public SelectedAutoCompleteSuggestionInfo(Range range, String text) {
    this.range = range;
    this.text = text;
  }
}
