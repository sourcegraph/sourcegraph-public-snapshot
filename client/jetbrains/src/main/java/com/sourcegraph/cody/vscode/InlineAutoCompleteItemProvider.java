package com.sourcegraph.cody.vscode;

import com.sourcegraph.cody.autocomplete.prompt_library.InlineAutoCompleteList;
import java.util.concurrent.CompletableFuture;

public abstract class InlineAutoCompleteItemProvider {
  public abstract CompletableFuture<InlineAutoCompleteList> provideInlineAutoCompleteItems(
      TextDocument document,
      Position position,
      InlineAutoCompleteContext context,
      CancellationToken token);
}
