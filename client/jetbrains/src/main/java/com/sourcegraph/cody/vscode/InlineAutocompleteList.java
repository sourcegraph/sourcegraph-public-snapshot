package com.sourcegraph.cody.vscode;

import com.sourcegraph.cody.agent.protocol.CompletionEvent;
import java.util.List;
import org.jetbrains.annotations.Nullable;

public class InlineAutocompleteList {
  @Nullable public List<InlineAutocompleteItem> items;
  @Nullable public CompletionEvent completionEvent;

  @Override
  public String toString() {
    return "InlineAutocompleteList{" + "items=" + items + '}';
  }
}
