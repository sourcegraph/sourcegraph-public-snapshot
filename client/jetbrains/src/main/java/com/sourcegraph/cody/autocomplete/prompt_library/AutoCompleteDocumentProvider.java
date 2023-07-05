package com.sourcegraph.cody.autocomplete.prompt_library;

import com.sourcegraph.cody.vscode.CancellationToken;
import com.sourcegraph.cody.vscode.Completion;
import com.sourcegraph.cody.vscode.TextDocumentContentProvider;
import java.net.URI;
import java.util.List;
import java.util.concurrent.CompletableFuture;

// TODO: implement the rest of this class.
public class AutoCompleteDocumentProvider implements TextDocumentContentProvider {
  @Override
  public CompletableFuture<String> provideTextDocumentContent(URI uri, CancellationToken token) {
    return null;
  }

  @Override
  public void clearCompletions(URI uri) {}

  @Override
  public void addCompletions(URI uri, String extension, List<Completion> completions) {}
}
