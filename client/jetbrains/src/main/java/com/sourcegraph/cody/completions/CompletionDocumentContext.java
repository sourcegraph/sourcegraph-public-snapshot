package com.sourcegraph.cody.completions;

public class CompletionDocumentContext {
  private final String sameLineSuffix;
  private final String sameLinePrefix;

  public CompletionDocumentContext(String sameLinePrefix, String sameLineSuffix) {
    this.sameLineSuffix = sameLineSuffix;
    this.sameLinePrefix = sameLinePrefix;
  }

  public String getSameLineSuffix() {
    return sameLineSuffix;
  }

  public String getSameLinePrefix() {
    return sameLinePrefix;
  }

  /**
   * We don't want to trigger completions when
   *
   * <ul>
   *   <li>the user is in the middle of a word
   *   <li>the suffix of the current line contains any word characters
   * </ul>
   *
   * @return whether it's valid to trigger a completion for the current document context.
   */
  public boolean isCompletionTriggerValid() {
    boolean prefixContainsText = sameLinePrefix.matches("\\s*[A-Za-z]+$");
    boolean suffixContainsWords = sameLineSuffix.matches(".*\\w.*");
    return !prefixContainsText && !suffixContainsWords;
  }

  @Override
  public String toString() {
    return "CompletionDocumentContext{"
        + "sameLineSuffix="
        + sameLineSuffix
        + ", sameLinePrefix="
        + sameLinePrefix
        + '}';
  }
}
