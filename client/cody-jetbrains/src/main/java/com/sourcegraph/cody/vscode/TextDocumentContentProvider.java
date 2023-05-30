package com.sourcegraph.cody.vscode;

import java.net.URI;
import java.util.List;
import java.util.concurrent.CompletableFuture;

public interface TextDocumentContentProvider {
  CompletableFuture<String> provideTextDocumentContent(URI uri, CancellationToken token);

  void clearCompletions(URI uri);

  void addCompletions(URI uri, String extension, List<Completion> completions);
}
