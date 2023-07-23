package com.sourcegraph.cody.autocomplete.prompt_library;

import static com.sourcegraph.cody.autocomplete.prompt_library.TextProcessing.*;

import java.util.regex.Pattern;

// This is mostly a rough line-by-line translation of the post-processing logic of the VS Code Cody
// plugin to Java.
// We will probably get rid of this in the future, but it's useful to have for now.
public class PostProcess {

  private static final Pattern newLineRegex = Pattern.compile("\n");

  public static String postProcess(String prefix, String rawResponse) {
    String completion = extractFromCodeBlock(rawResponse);

    boolean trimmedPrefixContainNewline =
        newLineRegex.matcher(prefix.substring(0, prefix.trim().length())).find();
    if (trimmedPrefixContainNewline) {
      completion = completion.stripLeading();
    } else {
      completion = trimStartUntilNewline(completion);
    }

    // Remove bad symbols from the start of the completion string.
    completion = fixBadCompletionStart(completion);

    // Remove incomplete lines in single-line completions
    int allowedNewlines = 2;
    String[] lines = completion.split("\n");
    if (lines.length >= allowedNewlines) {
      StringBuilder sb = new StringBuilder();
      for (int i = 0; i < allowedNewlines; i++) {
        sb.append(lines[i]);
        if (i < allowedNewlines - 1) {
          sb.append("\n");
        }
      }
      completion = sb.toString();
    }

    // Trim start and end of the completion to remove all trailing whitespace.
    return completion.trim();
  }
}
