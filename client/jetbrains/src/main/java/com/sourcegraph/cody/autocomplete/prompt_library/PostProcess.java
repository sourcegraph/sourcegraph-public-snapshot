package com.sourcegraph.cody.autocomplete.prompt_library;

import com.sourcegraph.cody.vscode.Completion;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.IntStream;

// This is mostly a rough line-by-line translation of the post-processing logic of the VS Code Cody
// plugin to Java.
// We will probably get rid of this in the future, but it's useful to have for now.
public class PostProcess {
  // using \p{So} instead of \p{Emoji_Presentation} because the latter does not have an equivalent
  // in Java; it seems some API for handling emojis will be added in JDK 21, or else we could
  // implement emoji handling ourselves if needed.
  // for now, let's use \p{So} which is the superset, catching all 'other symbols', including emojis
  private static final Pattern BAD_COMPLETION_START =
      Pattern.compile("^(\\p{So}|\\u200B|\\+ |- |. )+(\\s)+");

  public static Completion postProcess(
      String prefix,
      String suffix,
      //        String languageId, // TODO: this would be needed when we want to post-process
      // multiline completions
      //        boolean multiline,
      Completion completion) {

    String content = completion.content;

    // Extract a few common parts for the processing
    String currentLinePrefix = prefix.substring(prefix.lastIndexOf('\n') + 1);
    int firstNlInSuffix = suffix.indexOf('\n') + 1;
    String nextNonEmptyLine =
        Arrays.stream(suffix.substring(firstNlInSuffix).split("\n"))
            .filter(line -> line.trim().length() > 0)
            .findFirst()
            .orElse("");

    // Sometimes Claude emits a single space in the completion. We call this an "odd indentation"
    // completion and try to fix the response.
    boolean hasOddIndentation = false;
    if (content.length() > 0
        && content.startsWith(" ")
        && !content.startsWith("  ")
        && prefix.length() > 0
        && (prefix.endsWith(" ") || prefix.endsWith("\t"))) {
      content = content.substring(1);
      hasOddIndentation = true;
    }

    // Experimental: Trim start of the completion to remove all trailing whitespace nonsense
    content = content.stripLeading();

    // Detect bad completion start
    Matcher badCompletionStartMatcher = BAD_COMPLETION_START.matcher(content);
    if (badCompletionStartMatcher.find()) {
      content = content.substring(badCompletionStartMatcher.end());
    }

    // Strip out trailing markdown block and trim trailing whitespace
    int endBlockIndex = content.indexOf("```");
    if (endBlockIndex != -1) {
      content = content.substring(0, endBlockIndex);
    }

    //        if (multiline) { // TODO handle this part once we have multiline completions
    //            content = truncateMultilineCompletion(content, hasOddIndentation, prefix,
    // nextNonEmptyLine, languageId);
    //        } else
    if (content.contains("\n")) {
      content = content.substring(0, content.indexOf("\n"));
    }

    // If a completed line matches the next non-empty line of the suffix 1:1, we remove
    List<String> lines = Arrays.asList(content.split("\n"));
    Optional<Integer> matchedLineIndex =
        IntStream.range(0, lines.size())
            .boxed()
            .filter(
                i -> {
                  String line = lines.get(i);
                  if (i == 0) {
                    line = currentLinePrefix + line;
                  }
                  if (!line.trim().isEmpty() && !nextNonEmptyLine.trim().isEmpty()) {
                    // We need a trimEnd here because the machine likes to add trailing whitespace.
                    //
                    // TODO: Fix this earlier in the post process run but this needs to be careful
                    // not
                    // to alter the meaning
                    return line.trim().equals(nextNonEmptyLine);
                  }
                  return false;
                })
            .findFirst();
    if (matchedLineIndex.isPresent()) {
      content = String.join("\n", lines.subList(0, matchedLineIndex.get()));
    }

    return completion.withContent(content.trim());
  }

  //    private static String truncateMultilineCompletion(String content, boolean hasOddIndentation,
  // String prefix, String nextNonEmptyLine, String languageId) {
  //        return content; // placeholder // TODO implement/translate this part for multiline
  //    }
}
