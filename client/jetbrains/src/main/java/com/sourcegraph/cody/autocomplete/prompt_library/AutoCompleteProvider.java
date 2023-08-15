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

/**
 * Abstract base class for auto-complete providers. Subclasses must implement: createPromptPrefix()
 * - Generates prompt prefix messages. generateCompletions() - Generates completions for a given
 * context.
 */
public abstract class AutoCompleteProvider {
  /** The client for making completion requests */
  protected SourcegraphNodeCompletionsClient completionsClient;

  /** An executor service for async operations */
  protected final ExecutorService executor =
      Executors.newFixedThreadPool(CodyAutoCompleteItemProvider.nThreads);

  /** Max chars allowed for prompt */
  protected int promptChars;

  /** Max tokens to generate in response */
  protected int responseTokens;

  /** Reference code snippets to include in prompt */
  protected List<ReferenceSnippet> snippets;

  /** Prefix context before cursor */
  protected String prefix;

  /** Suffix context after cursor */
  protected String suffix;

  /** Additional prefix to inject into prompt */
  protected String injectPrefix;

  /** Default number of completions to generate */
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

  /** Gets length of prompt without snippets */
  public int emptyPromptLength() {
    String promptNoSnippets =
        createPromptPrefix().stream().map(Message::prompt).collect(Collectors.joining(""));
    return promptNoSnippets.length() - 10; // Hunch: The minus 10 is for "Assistant:"?
  }

  /** Builds full prompt with prefixes and snippets */
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

  /** Batches multiple completion requests */
  protected CompletableFuture<List<CompletionResponse>> batchCompletions(
      SourcegraphNodeCompletionsClient client, CompletionParameters params, int n) {
    List<CompletableFuture<CompletionResponse>> promises = new ArrayList<>();
    for (int i = 0; i < n; i++) {
      promises.add(client.complete(params));
    }
    return Promises.all(promises);
  }
}
