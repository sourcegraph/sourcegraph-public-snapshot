package com.sourcegraph.cody.completions.prompt_library;

import com.sourcegraph.cody.vscode.InlineCompletionItem;
import java.util.List;

public class InlineCompletionList {
  public final List<InlineCompletionItem> items;

  public InlineCompletionList(List<InlineCompletionItem> items) {
    this.items = items;
  }

  @Override
  public String toString() {
    return "InlineCompletionList{" + "items=" + items + '}';
  }
}
