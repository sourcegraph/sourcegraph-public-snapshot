package com.sourcegraph.cody.vscode;

import com.sourcegraph.cody.autocomplete.AutocompleteText;
import java.util.stream.Collectors;
import org.jetbrains.annotations.NotNull;

public class InlineAutocompleteItem {
  public final String insertText;
  public final String filterText;
  public final Range range;
  public final Command command;

  public InlineAutocompleteItem(
      String insertText, String filterText, Range range, Command command) {
    this.insertText = insertText;
    this.filterText = filterText;
    this.range = range;
    this.command = command;
  }

  public InlineAutocompleteItem withInsertText(String newInsertText) {
    return new InlineAutocompleteItem(newInsertText, this.filterText, this.range, this.command);
  }

  public InlineAutocompleteItem withFilterText(String newFilterText) {
    return new InlineAutocompleteItem(this.insertText, newFilterText, this.range, this.command);
  }

  public InlineAutocompleteItem withRange(Range newRange) {
    return new InlineAutocompleteItem(this.insertText, this.filterText, newRange, this.command);
  }

  public InlineAutocompleteItem withCommand(Command newCommand) {
    return new InlineAutocompleteItem(this.insertText, this.filterText, this.range, newCommand);
  }

  public static InlineAutocompleteItem fromCompletion(Completion completion) {
    return new InlineAutocompleteItem(
        completion.content,
        completion.prefix,
        new Range(new Position(0, 0), new Position(0, completion.content.length())),
        null);
  }

  public boolean isMultiline() {
    return this.insertText.lines().count() > 1;
  }

  @Override
  public String toString() {
    return "InlineAutocompleteItem{"
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

  public @NotNull AutocompleteText toAutocompleteText(@NotNull String sameLineSuffix) {
    boolean multiline = this.isMultiline();
    String sameLineRawAutocomplete =
        multiline ? this.insertText.lines().findFirst().orElse("") : this.insertText;
    boolean needAfterEndOfLineSuffix =
        !sameLineSuffix.isEmpty() && sameLineRawAutocomplete.contains(sameLineSuffix);
    int lastSuffixIndex = sameLineRawAutocomplete.lastIndexOf(sameLineSuffix);
    String sameLineBeforeSuffixText =
        needAfterEndOfLineSuffix
            ? sameLineRawAutocomplete.substring(0, lastSuffixIndex)
            : sameLineRawAutocomplete;
    String afterEndOfLineSuffix =
        needAfterEndOfLineSuffix
            ? sameLineRawAutocomplete.substring(lastSuffixIndex + sameLineSuffix.length())
            : "";
    String blockText =
        multiline
            ? this.insertText.lines().skip(1).collect(Collectors.joining(System.lineSeparator()))
            : "";
    return new AutocompleteText(sameLineBeforeSuffixText, afterEndOfLineSuffix, blockText);
  }
}
