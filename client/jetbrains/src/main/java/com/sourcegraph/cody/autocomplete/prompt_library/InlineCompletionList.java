package com.sourcegraph.cody.autocomplete.prompt_library;

import com.sourcegraph.cody.vscode.InlineAutoCompleteItem;
import java.util.List;

public class InlineCompletionList {
  public final List<InlineAutoCompleteItem> items;

  public InlineCompletionList(List<InlineAutoCompleteItem> items) {
    this.items = items;
  }

  @Override
  public String toString() {
    return "InlineCompletionList{" + "items=" + items + '}';
  }
}
