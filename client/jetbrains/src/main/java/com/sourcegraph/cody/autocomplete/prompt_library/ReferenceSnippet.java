package com.sourcegraph.cody.autocomplete.prompt_library;

public class ReferenceSnippet {
  public final String filename;
  public final JaccardMatch jaccard;

  public ReferenceSnippet(String filename, JaccardMatch jaccard) {
    this.filename = filename;
    this.jaccard = jaccard;
  }
}
