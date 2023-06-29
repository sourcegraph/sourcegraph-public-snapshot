package com.sourcegraph.cody.completions.prompt_library;

import com.sourcegraph.cody.api.Message;
import com.sourcegraph.cody.vscode.CancellationToken;
import com.sourcegraph.cody.vscode.Completion;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.CompletableFuture;

public class UnstableCodegenEndOfLineCompletionProvider extends EndOfLineCompletionProvider {
  public UnstableCodegenEndOfLineCompletionProvider(
      SourcegraphNodeCompletionsClient completionsClient,
      int promptChars,
      int responseTokens,
      List<ReferenceSnippet> snippets,
      String prefix,
      String suffix,
      String injectPrefix,
      int defaultN) {
    super(
        completionsClient,
        promptChars,
        responseTokens,
        snippets,
        prefix,
        suffix,
        injectPrefix,
        defaultN);
  }

  @Override
  protected List<Message> createPromptPrefix() {
    System.err.println("UnstableCodegenEndOfLineCompletionProvider.createPromptPrefix");
    // TODO: implement
    return List.of();
  }

  @Override
  public CompletableFuture<List<Completion>> generateCompletions(
      CancellationToken token, Optional<Integer> n) {
    System.err.println("UnstableCodegenEndOfLineCompletionProvider.generateCompletions");
    // TODO: implement
    return CompletableFuture.completedFuture(List.of());
  }
}
