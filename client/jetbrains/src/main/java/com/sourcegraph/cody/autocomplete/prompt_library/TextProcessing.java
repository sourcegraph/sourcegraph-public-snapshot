package com.sourcegraph.cody.autocomplete.prompt_library;

import java.util.Arrays;

public class TextProcessing {
  public static final String OPENING_CODE_TAG = "<CODE5711>";
  public static final String CLOSING_CODE_TAG = "</CODE5711>";

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
}
