package com.sourcegraph.cody.autocomplete.prompt_library;

import com.intellij.openapi.diagnostic.Logger;
import java.util.Arrays;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class TextProcessing {
  private static final Logger logger = Logger.getInstance(TextProcessing.class);
  public static final String OPENING_CODE_TAG = "<CODE5711>";
  public static final String CLOSING_CODE_TAG = "</CODE5711>";

  private static final Pattern closingCodeTagPattern =
      Pattern.compile(Pattern.quote(CLOSING_CODE_TAG));

  static class TrimmedString {
    String trimmed;
    String leadSpace;
    String rearSpace;

    public TrimmedString(String trimmed, String leadSpace, String rearSpace) {
      this.trimmed = trimmed;
      this.leadSpace = leadSpace;
      this.rearSpace = rearSpace;
    }
  }

  static class PrefixComponents {
    TrimmedString head;
    TrimmedString tail;
    String overlap;

    public PrefixComponents(TrimmedString head, TrimmedString tail, String overlap) {
      this.head = head;
      this.tail = tail;
      this.overlap = overlap;
    }
  }

  private static TrimmedString trimSpace(String s) {
    String trimmed = s.trim();
    int headEnd = s.indexOf(trimmed);
    String leadSpace = s.substring(0, headEnd);
    String rearSpace = s.substring(headEnd + trimmed.length());
    return new TrimmedString(trimmed, leadSpace, rearSpace);
  }

  public static PrefixComponents getHeadAndTail(String s) {
    String[] lines = s.split("\n");
    int tailThreshold = 2;

    int nonEmptyCount = 0;
    int tailStart = -1;
    for (int i = lines.length - 1; i >= 0; i--) {
      if (lines[i].trim().length() > 0) {
        nonEmptyCount++;
      }
      if (nonEmptyCount >= tailThreshold) {
        tailStart = i;
        break;
      }
    }

    if (tailStart == -1) {
      return new PrefixComponents(trimSpace(s), trimSpace(s), s);
    }

    String headLines = String.join("\n", Arrays.copyOfRange(lines, 0, tailStart));
    String tailLines = String.join("\n", Arrays.copyOfRange(lines, tailStart, lines.length));
    return new PrefixComponents(trimSpace(headLines), trimSpace(tailLines), null);
  }

  public static String extractFromCodeBlock(String completion) {
    if (completion.contains(OPENING_CODE_TAG)) {
      logger.warn(
          "Cody: invalid code completion response, should not contain opening tag <CODE5711>");
      return "";
    }

    String[] splitCompletion = closingCodeTagPattern.split(completion, 0);
    ;
    String result = (splitCompletion.length > 0) ? splitCompletion[0] : "";

    return result.trim();
  }

  // using \p{So} instead of \p{Emoji_Presentation} because the latter does not have an equivalent
  // in Java; it seems some API for handling emojis will be added in JDK 21, or else we could
  // implement emoji handling ourselves if needed.
  // for now, let's use \p{So} which is the superset, catching all 'other symbols', including emojis
  private static final Pattern BAD_COMPLETION_START =
      Pattern.compile("^(\\p{So}|\\u200B|\\+ |- |\\. )+(\\s)+");

  public static String fixBadCompletionStart(String completion) {
    Matcher matcher = BAD_COMPLETION_START.matcher(completion);
    if (matcher.find()) return completion.replaceFirst(BAD_COMPLETION_START.pattern(), "");
    else return completion;
  }

  public static String trimStartUntilNewline(String str) {
    int index = str.indexOf('\n');
    if (index == -1) return str.replaceAll("^\\s+", "");

    String firstPart = str.substring(0, index).replaceAll("^\\s+", "");
    String secondPart = str.substring(index);
    return firstPart + secondPart;
  }
}
