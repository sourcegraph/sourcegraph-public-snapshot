package com.sourcegraph.cody.completions.prompt_library;

public class CompletionResponse {
  public final String completion;
  public final String stopReason;

  public CompletionResponse(String completion, String stopReason) {
    this.completion = completion;
    this.stopReason = stopReason;
  }
}
