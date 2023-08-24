package com.sourcegraph.cody.vscode;

import java.util.List;

public class InlineAutocompleteList {
  public final List<InlineAutocompleteItem> items;

  public InlineAutocompleteList(List<InlineAutocompleteItem> items) {
    this.items = items;
  }

  @Override
  public String toString() {
    return "InlineAutocompleteList{" + "items=" + items + '}';
  }
}
