package com.sourcegraph.cody.vscode;

import com.sourcegraph.cody.autocomplete.prompt_library.InlineAutoCompleteList;
import java.util.concurrent.CompletableFuture;

public abstract class InlineCompletionItemProvider {
  public abstract CompletableFuture<InlineAutoCompleteList> provideInlineCompletions(
      TextDocument document,
      Position position,
      InlineCompletionContext context,
      CancellationToken token);
}
