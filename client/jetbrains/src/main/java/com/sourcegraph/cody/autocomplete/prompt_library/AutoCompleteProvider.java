package com.sourcegraph.cody.autocomplete.prompt_library;

import static com.sourcegraph.cody.autocomplete.prompt_library.TextProcessing.CLOSING_CODE_TAG;
import static com.sourcegraph.cody.autocomplete.prompt_library.TextProcessing.OPENING_CODE_TAG;

import com.sourcegraph.cody.api.Message;
import com.sourcegraph.cody.api.Promises;
import com.sourcegraph.cody.api.Speaker;
import com.sourcegraph.cody.vscode.CancellationToken;
import com.sourcegraph.cody.vscode.Completion;
import java.util.*;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.stream.Collectors;

public abstract class AutoCompleteProvider {
  protected SourcegraphNodeCompletionsClient completionsClient;
  protected final ExecutorService executor =
      Executors.newFixedThreadPool(CodyAutoCompleteItemProvider.nThreads);
  protected int promptChars;
  protected int responseTokens;
  protected List<ReferenceSnippet> snippets;
  protected String prefix;
  protected String suffix;
  protected String injectPrefix;
  protected int defaultN;

  public AutoCompleteProvider(
      SourcegraphNodeCompletionsClient completionsClient,
      int promptChars,
      int responseTokens,
      List<ReferenceSnippet> snippets,
      String prefix,
      String suffix,
      String injectPrefix,
      int defaultN) {
    this.completionsClient = completionsClient;
    this.promptChars = promptChars;
    this.responseTokens = responseTokens;
    this.snippets = snippets;
    this.prefix = prefix;
    this.suffix = suffix;
    this.injectPrefix = injectPrefix;
    this.defaultN = defaultN;
  }

  protected abstract List<Message> createPromptPrefix();

  @SuppressWarnings("OptionalUsedAsFieldOrParameterType")
  public abstract CompletableFuture<List<Completion>> generateCompletions(
      CancellationToken token, Optional<Integer> n);

  public int emptyPromptLength() {
    String promptNoSnippets =
        createPromptPrefix().stream().map(Message::prompt).collect(Collectors.joining(""));
    return promptNoSnippets.length() - 10;
  }

  protected List<Message> createPrompt() {
    List<Message> prefixMessages = createPromptPrefix();
    List<Message> referenceSnippetMessages = new ArrayList<>();

    int remainingChars = promptChars - emptyPromptLength();

    for (ReferenceSnippet snippet : snippets) {
      List<Message> snippetMessages =
          List.of(
              new Message(
                  Speaker.HUMAN,
                  "Here is a reference snippet of code: "
                      + OPENING_CODE_TAG
                      + snippet.jaccard.text
                      + CLOSING_CODE_TAG),
              new Message(
                  Speaker.ASSISTANT, "Okay, I have added the snippet to my knowledge base."));
      int numSnippetChars =
          snippetMessages.stream().map(Message::prompt).collect(Collectors.joining("")).length()
              + 1;
      if (numSnippetChars > remainingChars) {
        break;
      }
      referenceSnippetMessages.addAll(snippetMessages);
      remainingChars -= numSnippetChars;
    }

    referenceSnippetMessages.addAll(prefixMessages);
    return referenceSnippetMessages;
  }

  protected CompletableFuture<List<CompletionResponse>> batchCompletions(
      SourcegraphNodeCompletionsClient client, CompletionParameters params, int n) {
    List<CompletableFuture<CompletionResponse>> promises = new ArrayList<>();
    for (int i = 0; i < n; i++) {
      promises.add(client.complete(params));
    }
    return Promises.all(promises);
  }
}
