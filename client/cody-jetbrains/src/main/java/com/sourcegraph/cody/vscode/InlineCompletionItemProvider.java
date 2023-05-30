package com.sourcegraph.cody.vscode;

import com.sourcegraph.cody.completions.prompt_library.InlineCompletionList;
import java.util.concurrent.CompletableFuture;

public abstract class InlineCompletionItemProvider {
  public abstract CompletableFuture<InlineCompletionList> provideInlineCompletions(
      TextDocument document,
      Position position,
      InlineCompletionContext context,
      CancellationToken token);
}
