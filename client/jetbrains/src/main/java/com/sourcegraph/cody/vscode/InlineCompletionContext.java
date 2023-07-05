package com.sourcegraph.cody.vscode;

public class InlineCompletionContext {
  public final InlineCompletionTriggerKind triggerKind;
  public final SelectedAutoCompleteSuggestionInfo selectedAutoCompleteSuggestionInfo;

  public InlineCompletionContext(
      InlineCompletionTriggerKind triggerKind,
      SelectedAutoCompleteSuggestionInfo selectedAutoCompleteSuggestionInfo) {
    this.triggerKind = triggerKind;
    this.selectedAutoCompleteSuggestionInfo = selectedAutoCompleteSuggestionInfo;
  }
}
