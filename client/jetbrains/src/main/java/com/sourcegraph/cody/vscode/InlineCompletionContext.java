package com.sourcegraph.cody.vscode;

public class InlineCompletionContext {
  public final InlineCompletionTriggerKind triggerKind;
  public final SelectedCompletionInfo selectedCompletionInfo;

  public InlineCompletionContext(
      InlineCompletionTriggerKind triggerKind, SelectedCompletionInfo selectedCompletionInfo) {
    this.triggerKind = triggerKind;
    this.selectedCompletionInfo = selectedCompletionInfo;
  }
}
