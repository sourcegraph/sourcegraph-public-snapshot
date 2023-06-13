package com.sourcegraph.cody.vscode;

public class SelectedCompletionInfo {
  public final Range range;
  public final String text;

  public SelectedCompletionInfo(Range range, String text) {
    this.range = range;
    this.text = text;
  }
}
