package com.sourcegraph.cody.vscode;

import java.util.concurrent.CompletableFuture;

public abstract class InlineAutocompleteItemProvider {
  public abstract CompletableFuture<InlineAutocompleteList> provideInlineAutocompleteItems(
      TextDocument document,
      Position position,
      InlineAutocompleteContext context,
      CancellationToken token);
}
