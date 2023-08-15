package com.sourcegraph.cody.autocomplete.prompt_library;

import static com.sourcegraph.cody.autocomplete.prompt_library.TextProcessing.*;

import com.intellij.openapi.diagnostic.Logger;
import com.sourcegraph.cody.api.Message;
import com.sourcegraph.cody.api.Speaker;
import com.sourcegraph.cody.vscode.CancellationToken;
import com.sourcegraph.cody.vscode.Completion;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.CompletableFuture;
import java.util.stream.Collectors;

public class AnthropicAutoCompleteProvider extends AutoCompleteProvider {

  public static final Logger logger = Logger.getInstance(AnthropicAutoCompleteProvider.class);

  public AnthropicAutoCompleteProvider(
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
    String[] prefixLines = this.prefix.split("\n");
    if (prefixLines.length == 0) logger.warn("Cody: Anthropic: missing prefix lines");

    PrefixComponents pc = getHeadAndTail(this.prefix);

    return Arrays.asList(
        new Message(
            Speaker.HUMAN,
            "You are Cody, a code completion AI developed by Sourcegraph. You write code in between tags like this:"
                + OPENING_CODE_TAG
                + "/* Code goes here */"
                + CLOSING_CODE_TAG),
        new Message(Speaker.ASSISTANT, "I am Cody, a code completion AI developed by Sourcegraph."),
        new Message(
            Speaker.HUMAN,
            "Complete this code: " + OPENING_CODE_TAG + pc.head.trimmed + CLOSING_CODE_TAG + "."),
        new Message(
            Speaker.ASSISTANT, "Okay, here is some code: " + OPENING_CODE_TAG + pc.tail.trimmed));
  }

  @Override
  public CompletableFuture<List<Completion>> generateCompletions(
      CancellationToken abortSignal, Optional<Integer> n) {
    String prefix = this.prefix + this.injectPrefix;

    // Create prompt
    List<Message> prompt = this.createPrompt();
    if (prompt.size() > this.promptChars) {
      logger.warn("Cody: Anthropic: prompt length exceeded maximum allotted chars");
      return CompletableFuture.completedFuture(Collections.emptyList());
    }

    // Issue request
    int maxTokensToSample = Math.min(100, this.responseTokens);
    List<String> stopSequences = List.of(Speaker.HUMAN.prompt(), CLOSING_CODE_TAG, "\n\n");
    CompletableFuture<List<CompletionResponse>> promises =
        batchCompletions(
            this.completionsClient,
            new CompletionParameters()
                .withMessages(prompt)
                .withMaxTokensToSample(maxTokensToSample)
                .withStopSequences(stopSequences)
                .withTemperature(0.5f)
                .withTopK(-1)
                .withTopP(-1),
            n.orElseGet(() -> this.defaultN));

    // Post-process
    return promises.thenApply(
        responses ->
            responses.stream()
                .map(
                    resp ->
                        new Completion(
                            prefix,
                            prompt,
                            PostProcess.postProcess(this.prefix, resp.completion),
                            resp.stopReason))
                .collect(Collectors.toList()));
  }
}
