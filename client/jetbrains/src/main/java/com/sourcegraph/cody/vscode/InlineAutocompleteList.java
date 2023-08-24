package com.sourcegraph.cody.vscode;

import java.util.List;

public class InlineAutocompleteList {
  public final List<InlineAutocompleteItem> items;
  public final CompletionEvent completionEvent;

  public InlineAutocompleteList(List<InlineAutocompleteItem> items, CompletionEvent completionEvent) {
    this.items = items;
    this.completionEvent = completionEvent;
  }

  @Override
  public String toString() {
    return "InlineAutocompleteList{" + "items=" + items + ", competionEvent=" + completionEvent + '}';
  }
}
