package com.sourcegraph.cody.vscode;

public class InlineCompletionItem {
  public final String insertText;
  public final String filterText;
  public final Range range;
  public final Command command;

  public InlineCompletionItem(String insertText, String filterText, Range range, Command command) {
    this.insertText = insertText;
    this.filterText = filterText;
    this.range = range;
    this.command = command;
  }

  public static InlineCompletionItem fromCompletion(Completion completion) {
    return new InlineCompletionItem(
        completion.content,
        completion.prefix,
        new Range(new Position(0, 0), new Position(0, completion.content.length())),
        null);
  }

  @Override
  public String toString() {
    return "InlineCompletionItem{"
        + "insertText='"
        + insertText
        + '\''
        + ", filterText='"
        + filterText
        + '\''
        + ", range="
        + range
        + ", command="
        + command
        + '}';
  }
}
