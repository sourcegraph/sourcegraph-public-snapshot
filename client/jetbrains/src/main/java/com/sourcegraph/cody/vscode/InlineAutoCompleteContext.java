package com.sourcegraph.cody.vscode;

public class InlineAutoCompleteContext {
  public final InlineAutoCompleteTriggerKind triggerKind;
  public final SelectedAutoCompleteSuggestionInfo selectedAutoCompleteSuggestionInfo;

  public InlineAutoCompleteContext(
      InlineAutoCompleteTriggerKind triggerKind,
      SelectedAutoCompleteSuggestionInfo selectedAutoCompleteSuggestionInfo) {
    this.triggerKind = triggerKind;
    this.selectedAutoCompleteSuggestionInfo = selectedAutoCompleteSuggestionInfo;
  }
}
