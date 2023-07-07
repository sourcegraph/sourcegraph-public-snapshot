package com.sourcegraph.cody.autocomplete.prompt_library;

public class JaccardMatch {
  public final int score;
  public final String text;

  public JaccardMatch(int score, String text) {
    this.score = score;
    this.text = text;
  }
}
