package com.sourcegraph.cody.autocomplete.prompt_library;

public class DocContext {
  public final String prefix;
  public final String suffix;
  public final String prevLine;
  public final String prevNonEmptyLine;
  public final String nextNonEmptyLine;

  public DocContext(
      String prefix,
      String suffix,
      String prevLine,
      String prevNonEmptyLine,
      String nextNonEmptyLine) {
    this.prefix = prefix;
    this.suffix = suffix;
    this.prevLine = prevLine;
    this.prevNonEmptyLine = prevNonEmptyLine;
    this.nextNonEmptyLine = nextNonEmptyLine;
  }
}
