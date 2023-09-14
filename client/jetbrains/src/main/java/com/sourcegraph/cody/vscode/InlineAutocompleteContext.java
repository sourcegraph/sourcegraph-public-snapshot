package com.sourcegraph.cody.vscode;

public class InlineAutocompleteContext {
  public final InlineAutocompleteTriggerKind triggerKind;
  public final SelectedAutocompleteSuggestionInfo selectedAutocompleteSuggestionInfo;

  public InlineAutocompleteContext(
      InlineAutocompleteTriggerKind triggerKind,
      SelectedAutocompleteSuggestionInfo selectedAutocompleteSuggestionInfo) {
    this.triggerKind = triggerKind;
    this.selectedAutocompleteSuggestionInfo = selectedAutocompleteSuggestionInfo;
  }
}
