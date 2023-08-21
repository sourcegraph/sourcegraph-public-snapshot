package com.sourcegraph.cody.vscode;

import java.util.List;

public class InlineAutoCompleteList {
  public final List<InlineAutoCompleteItem> items;

  public InlineAutoCompleteList(List<InlineAutoCompleteItem> items) {
    this.items = items;
  }

  @Override
  public String toString() {
    return "InlineAutoCompleteList{" + "items=" + items + '}';
  }
}
