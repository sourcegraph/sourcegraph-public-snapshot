package com.sourcegraph.cody.vscode;

import java.util.concurrent.CompletableFuture;

public abstract class InlineAutoCompleteItemProvider {
  public abstract CompletableFuture<InlineAutoCompleteList> provideInlineAutoCompleteItems(
      TextDocument document,
      Position position,
      InlineAutoCompleteContext context,
      CancellationToken token);
}
