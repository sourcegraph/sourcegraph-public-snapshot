package com.sourcegraph.cody.autocomplete.prompt_library;

import com.sourcegraph.cody.api.Message;
import com.sourcegraph.cody.api.Speaker;
import com.sourcegraph.cody.vscode.CancellationToken;
import com.sourcegraph.cody.vscode.Completion;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.CompletableFuture;
import java.util.stream.Collectors;

public class EndOfLineAutoCompleteProvider extends AutoCompleteProvider {
  public EndOfLineAutoCompleteProvider(
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
    if (prefixLines.length == 0) {
      throw new Error("no prefix lines");
    }

    List<Message> prefixMessages;
    if (prefixLines.length > 2) {
      int endLine = Math.max(prefixLines.length / 2, prefixLines.length - 5);
      prefixMessages =
          List.of(
              new Message(
                  Speaker.HUMAN,
                  "Complete the following file:\n"
                      + "```\n"
                      + String.join("\n", Arrays.copyOfRange(prefixLines, 0, endLine))
                      + "\n"
                      + "```"),
              new Message(
                  Speaker.ASSISTANT,
                  "Here is the completion of the file:\n"
                      + "```\n"
                      + String.join(
                          "\n", Arrays.copyOfRange(prefixLines, endLine, prefixLines.length))
                      + this.injectPrefix));
    } else {
      prefixMessages =
          List.of(
              new Message(Speaker.HUMAN, "Write some code"),
              new Message(
                  Speaker.ASSISTANT,
                  "Here is some code:\n```\n" + this.prefix + this.injectPrefix + "```"));
    }

    return prefixMessages;
  }

  @Override
  public CompletableFuture<List<Completion>> generateCompletions(
      CancellationToken abortSignal, Optional<Integer> n) {
    String prefix = this.prefix + this.injectPrefix;

    // Create prompt
    List<Message> prompt = this.createPrompt();
    if (prompt.size() > this.promptChars) {
      throw new Error("prompt length exceeded maximum alloted chars");
    }

    // Issue request
    CompletableFuture<List<CompletionResponse>> promises =
        batchCompletions(
            this.completionsClient,
            new CompletionParameters()
                .withMessages(prompt)
                .withMaxTokensToSample(this.responseTokens)
                .withStopSequences(List.of(Speaker.HUMAN.prompt(), "\n"))
                .withTemperature(1)
                .withTopK(-1)
                .withTopP(-1),
            n.isEmpty() ? this.defaultN : n.get());

    // Post-process
    return promises.thenApply(
        responses ->
            responses.stream()
                .map(
                    resp ->
                        new Completion(
                            prefix, prompt, this.postProcess(resp.completion), resp.stopReason))
                .collect(Collectors.toList()));
  }

  private String postProcess(String completion) {
    // Sometimes Claude emits an extra space
    if (completion.startsWith(" ") && this.prefix.endsWith(" ")) {
      completion = completion.substring(1);
    }
    // Insert the injected prefix back in
    if (this.injectPrefix.length() > 0) {
      completion = this.injectPrefix + completion;
    }
    // Strip out trailing markdown block and trim trailing whitespace
    int endBlockIndex = completion.indexOf("```");
    if (endBlockIndex != -1) {
      return completion.substring(0, endBlockIndex).trim();
    }
    return completion.trim();
  }
}
