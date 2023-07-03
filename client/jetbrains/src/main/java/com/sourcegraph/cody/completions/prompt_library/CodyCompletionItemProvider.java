package com.sourcegraph.cody.completions.prompt_library;

import com.sourcegraph.cody.api.Promises;
import com.sourcegraph.cody.completions.CompletionsProviderType;
import com.sourcegraph.cody.vscode.*;
import com.sourcegraph.config.UserLevelConfig;
import java.util.*;
import java.util.concurrent.*;
import java.util.function.Function;
import java.util.stream.Collectors;

/**
 * Manually translated logic from <code>client/cody/src/completions/index.ts</code> in the VS Code
 * extension. Some code in this class is not used since we haven't translated all the logic yet.
 * Let's keep the unused code to make it easier to see the similarity between the two versions.
 */
@SuppressWarnings({"unused", "FieldCanBeLocal", "CommentedOutCode"})
public class CodyCompletionItemProvider extends InlineCompletionItemProvider {
  int nThreads = 3; // up to 3 completion API calls to run in parallel
  // should we reuse the scheduler from CodyCompletionsManager here later on?
  private final ExecutorService executor = Executors.newFixedThreadPool(nThreads);
  private final int promptTokens;
  private final int maxPrefixTokens;
  private final int maxSuffixTokens;
  private Runnable abortOpenInlineCompletions = () -> {};
  private final Runnable abortOpenMultilineCompletion = () -> {};
  private final WebviewErrorMessenger webviewErrorMessenger;
  private final SourcegraphNodeCompletionsClient completionsClient;
  private final CompletionsDocumentProvider documentProvider;
  private final History history;
  private final int charsPerToken;
  private final int responseTokens;

  public CodyCompletionItemProvider(
      WebviewErrorMessenger webviewErrorMessenger,
      SourcegraphNodeCompletionsClient completionsClient,
      CompletionsDocumentProvider documentProvider,
      History history,
      int contextWindowTokens,
      int charsPerToken,
      int responseTokens,
      double prefixPercentage,
      double suffixPercentage) {
    this.webviewErrorMessenger = webviewErrorMessenger;
    this.completionsClient = completionsClient;
    this.documentProvider = documentProvider;
    this.history = history;
    this.charsPerToken = charsPerToken;
    this.responseTokens = responseTokens;
    this.promptTokens = contextWindowTokens - responseTokens;
    this.maxPrefixTokens = (int) Math.floor(promptTokens * prefixPercentage);
    this.maxSuffixTokens = (int) Math.floor(promptTokens * suffixPercentage);
  }

  @Override
  public CompletableFuture<InlineCompletionList> provideInlineCompletions(
      TextDocument document,
      Position position,
      InlineCompletionContext context,
      CancellationToken token) {
    try {
      return provideInlineCompletionItemsInner(document, position, context, token);
    } catch (Exception e) {
      if (e.getMessage().equals("aborted")) {
        return emptyResult();
      }
      return emptyResult();
    }
  }

  private CompletableFuture<InlineCompletionList> emptyResult() {
    return CompletableFuture.completedFuture(new InlineCompletionList(List.of()));
  }

  private int tokToChar(int toks) {
    return toks * charsPerToken;
  }

  private CompletableFuture<InlineCompletionList> provideInlineCompletionItemsInner(
      TextDocument document,
      Position position,
      InlineCompletionContext context,
      CancellationToken token) {
    this.abortOpenInlineCompletions.run();
    CancellationToken abortController = new CancellationToken();
    token.onCancellationRequested(abortController::abort);
    this.abortOpenInlineCompletions = abortController::abort;

    DocContext docContext =
        getCurrentDocContext(
            document, position, tokToChar(maxPrefixTokens), tokToChar(maxSuffixTokens));
    if (docContext == null) {
      return emptyResult();
    }

    String prefix = docContext.prefix;
    String suffix = docContext.suffix;
    String precedingLine = docContext.prevLine;

    // TODO: implement caching
    //    Completion[] cachedCompletions = inlineCompletionsCache.get(prefix);
    //    if (cachedCompletions != null) {
    //      return cachedCompletions.stream()
    //          .map(InlineCompletionItem::new)
    //          .toArray(InlineCompletionItem[]::new);
    //    }

    int remainingChars = tokToChar(promptTokens);

    CompletionProvider completionNoSnippets =
        endOfLineProvider(
            completionsClient,
            remainingChars,
            responseTokens,
            List.of(),
            prefix,
            suffix,
            "\n",
            1,
            document);
    int emptyPromptLength = completionNoSnippets.emptyPromptLength();

    List<ReferenceSnippet> similarCode = Collections.emptyList();

    //    int waitMs;
    List<CompletionProvider> completers = new ArrayList<>();

    if (context.selectedCompletionInfo != null || precedingLine.matches(".*[A-Za-z]$")) {
      return emptyResult();
    }

    if (precedingLine.trim().equals("")) {
      //      waitMs = 500;
      completers.add(
          endOfLineProvider(
              completionsClient,
              remainingChars,
              responseTokens,
              similarCode,
              prefix,
              suffix,
              "",
              2,
              document));
    } else if (context.triggerKind == InlineCompletionTriggerKind.Invoke
        || precedingLine.endsWith(".")) {
      return emptyResult();
    } else {
      //      waitMs = 1000;
      completers.add(
          endOfLineProvider(
              completionsClient,
              remainingChars,
              responseTokens,
              similarCode,
              prefix,
              suffix,
              "",
              2,
              document));
      completers.add(
          endOfLineProvider(
              completionsClient,
              remainingChars,
              responseTokens,
              similarCode,
              prefix,
              suffix,
              "\n",
              1,
              document));
    }

    // TODO: implement debouncing with a non-blocking way instead of `Thread.sleep()`
    //    try {
    //      Thread.sleep(waitMs);
    //    } catch (InterruptedException e) {
    //      throw new RuntimeException(e);
    //    }

    if (abortController.isCancelled()) {
      return emptyResult();
    }
    List<CompletableFuture<List<Completion>>> promises =
        completers.stream()
            .map(
                c -> // ensure completions are called for asynchronously
                CompletableFuture.supplyAsync(
                        () -> c.generateCompletions(token, Optional.empty()), executor))
            .map(cf -> cf.thenCompose(Function.identity())) // flatten the CompletableFutures
            .collect(Collectors.toList());
    CompletableFuture<List<InlineCompletionItem>> all =
        Promises.all(promises)
            .thenApply(
                completions ->
                    completions.stream()
                        .flatMap(Collection::stream)
                        .map(c -> PostProcess.postProcess(prefix, suffix, c))
                        .map(InlineCompletionItem::fromCompletion)
                        .collect(Collectors.toList()));

    return all.thenApply(InlineCompletionList::new);
  }

  private DocContext getCurrentDocContext(
      TextDocument document, Position position, int maxPrefixLength, int maxSuffixLength) {
    int offset = document.offsetAt(position);
    String[] prefixLines = document.getText(new Range(new Position(0, 0), position)).split("\n");
    if (prefixLines.length == 0) {
      return null;
    }

    String[] suffixLines =
        document
            .getText(new Range(position, document.positionAt(document.getText().length())))
            .split("\n");
    String nextNonEmptyLine = "";
    if (suffixLines.length > 0) {
      for (String line : suffixLines) {
        if (line.trim().length() > 0) {
          nextNonEmptyLine = line;
          break;
        }
      }
    }

    String prevNonEmptyLine = "";
    for (int i = prefixLines.length - 1; i >= 0; i--) {
      String line = prefixLines[i];
      if (line.trim().length() > 0) {
        prevNonEmptyLine = line;
        break;
      }
    }

    String prevLine = prefixLines[prefixLines.length - 1];

    String prefix;
    if (offset > maxPrefixLength) {
      int total = 0;
      int startLine = prefixLines.length;
      for (int i = prefixLines.length - 1; i >= 0; i--) {
        if (total + prefixLines[i].length() > maxPrefixLength) {
          break;
        }
        startLine = i;
        total += prefixLines[i].length();
      }
      prefix = String.join("\n", Arrays.copyOfRange(prefixLines, startLine, prefixLines.length));
    } else {
      prefix = document.getText(new Range(new Position(0, 0), position));
    }

    int totalSuffix = 0;
    int endLine = 0;
    for (int i = 0; i < suffixLines.length; i++) {
      if (totalSuffix + suffixLines[i].length() > maxSuffixLength) {
        break;
      }
      endLine = i + 1;
      totalSuffix += suffixLines[i].length();
    }
    String suffix = String.join("\n", Arrays.copyOfRange(suffixLines, 0, endLine));

    return new DocContext(prefix, suffix, prevLine, prevNonEmptyLine, nextNonEmptyLine);
  }

  private CompletionProvider endOfLineProvider(
      SourcegraphNodeCompletionsClient completionsClient,
      int promptChars,
      int responseTokens,
      List<ReferenceSnippet> snippets,
      String prefix,
      String suffix,
      String injectPrefix,
      int defaultN,
      TextDocument document) {
    Function<Optional<String>, CompletionProvider> fallbackDefaultProvider =
        (maybeErrorToLog) -> {
          maybeErrorToLog.ifPresent(System.err::println);
          return new EndOfLineCompletionProvider(
              completionsClient,
              promptChars,
              responseTokens,
              snippets,
              prefix,
              suffix,
              injectPrefix,
              defaultN);
        };
    CompletionsProviderType providerType = UserLevelConfig.getCompletionsProviderType();
    Optional<String> completionsServerEndpoint =
        Optional.ofNullable(UserLevelConfig.getCompletionsServerEndpoint());
    if (providerType == CompletionsProviderType.UNSTABLE_CODEGEN) {
      return completionsServerEndpoint
          .map(
              endpoint ->
                  (CompletionProvider)
                      new UnstableCodegenEndOfLineCompletionProvider(
                          snippets,
                          prefix,
                          suffix,
                          document.fileName(),
                          endpoint,
                          document.getLanguageId()))
          .orElseGet(
              () ->
                  fallbackDefaultProvider.apply(
                      Optional.of(
                          "Error: Cody: missing completions server endpoint, falling back to anthropic completions provider")));
    } else return fallbackDefaultProvider.apply(Optional.empty());
  }
}
