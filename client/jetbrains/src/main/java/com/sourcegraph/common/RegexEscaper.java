package com.sourcegraph.common;

import java.util.regex.Pattern;
import org.jetbrains.annotations.NotNull;

public class RegexEscaper {
  static final Pattern SPECIAL_REGEX_CHARS = Pattern.compile("[{}()\\[\\].+*?^$\\\\|]");

  @NotNull
  public static String escapeRegexChars(@NotNull String string) {
    return SPECIAL_REGEX_CHARS.matcher(string).replaceAll("\\\\$0");
  }
}
