package com.sourcegraph.cody.autocomplete.prompt_library;

import com.sourcegraph.cody.api.Message;
import com.sourcegraph.cody.api.Speaker;
import com.sourcegraph.cody.vscode.CancellationToken;
import com.sourcegraph.cody.vscode.Completion;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.CompletableFuture;
import java.util.stream.Collectors;

public class MultilineAutoCompleteProvider extends AutoCompleteProvider {
  public MultilineAutoCompleteProvider(
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

  protected List<Message> createPromptPrefix() {
    // TODO(beyang): escape 'Human:' and 'Assistant:'
    String prefix = this.prefix.trim();

    String[] prefixLines = prefix.split("\n");
    if (prefixLines.length == 0) {
      throw new Error("no prefix lines");
    }

    List<Message> prefixMessages = new ArrayList<>();
    if (prefixLines.length > 2) {
      int endLine = Math.max(prefixLines.length / 2, prefixLines.length - 5);
      prefixMessages.add(
          new Message(
              Speaker.HUMAN,
              "Complete the following file:\n"
                  + "```\n"
                  + String.join("\n", Arrays.copyOfRange(prefixLines, 0, endLine))
                  + "\n"
                  + "```"));
      prefixMessages.add(
          new Message(
              Speaker.ASSISTANT,
              "Here is the completion of the file:\n```\n"
                  + String.join(
                      "\n", Arrays.copyOfRange(prefixLines, endLine, prefixLines.length))));
    } else {
      prefixMessages.add(new Message(Speaker.HUMAN, "Write some code"));
      prefixMessages.add(
          new Message(Speaker.ASSISTANT, "Here is some code:\n```\n" + prefix + "```"));
    }

    return prefixMessages;
  }

  private String postProcess(String completion) {
    String suggestion = completion;
    int endBlockIndex = completion.indexOf("```");
    if (endBlockIndex != -1) {
      suggestion = completion.substring(0, endBlockIndex);
    }

    // Remove trailing whitespace before newlines
    String suggestionWithoutTrailingWhitespace =
        Arrays.stream(suggestion.split("\n")).map(String::trim).collect(Collectors.joining("\n"));

    return sliceUntilFirstNLinesOfSuffixMatch(suggestionWithoutTrailingWhitespace, this.suffix, 5);
  }

  @Override
  public CompletableFuture<List<Completion>> generateCompletions(
      CancellationToken abortSignal, Optional<Integer> n) {
    String prefix = this.prefix.trim();

    // Create prompt
    List<Message> prompt = this.createPrompt();
    String textPrompt = prompt.stream().map(Message::prompt).collect(Collectors.joining(""));
    if (textPrompt.length() > this.promptChars) {
      throw new Error("prompt length exceeded maximum alloted chars");
    }

    // Issue request
    CompletableFuture<List<CompletionResponse>> promises =
        batchCompletions(
            this.completionsClient,
            new CompletionParameters()
                .withMessages(prompt)
                .withMaxTokensToSample(this.responseTokens),
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

  /**
   * This function slices the suggestion string until the first n lines match the suffix string.
   *
   * <p>It splits suggestion and suffix into lines, then iterates over the lines of suffix. For each
   * line of suffix, it checks if the next n lines of suggestion match. If so, it returns the first
   * part of suggestion up to those matching lines. If no match is found after iterating all lines
   * of suffix, the full suggestion is returned.
   *
   * <p>For example, with: suggestion = "foo\nbar\nbaz\nqux\nquux" suffix = "baz\nqux\nquux" n = 3
   *
   * <p>It would return: "foo\nbar"
   *
   * <p>Because the first 3 lines of suggestion ("baz\nqux\nquux") match suffix.
   */
  public static String sliceUntilFirstNLinesOfSuffixMatch(String suggestion, String suffix, int n) {
    String[] suggestionLines = suggestion.split("\n");
    String[] suffixLines = suffix.split("\n");

    for (int i = 0; i < suffixLines.length; i++) {
      int matchedLines = 0;
      for (int j = 0; j < suggestionLines.length; j++) {
        if (suffixLines.length < i + matchedLines) {
          continue;
        }
        if (suffixLines[i + matchedLines].equals(suggestionLines[j])) {
          matchedLines++;
        } else {
          matchedLines = 0;
        }
        if (matchedLines >= n) {
          return String.join("\n", Arrays.copyOfRange(suggestionLines, 0, j - n + 1));
        }
      }
    }

    return suggestion;
  }
}
