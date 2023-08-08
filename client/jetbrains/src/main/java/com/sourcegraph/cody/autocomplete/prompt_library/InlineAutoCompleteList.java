package com.sourcegraph.cody.autocomplete.prompt_library;

import com.sourcegraph.cody.vscode.InlineAutoCompleteItem;
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
