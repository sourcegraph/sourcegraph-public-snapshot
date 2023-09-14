package com.sourcegraph.cody.api;

import java.util.List;
import org.jetbrains.annotations.NotNull;

/** Input for the completions request. */
public class CompletionsInput {
  public @NotNull List<Message> messages;
  public float temperature;
  public List<String> stopSequences;
  public int maxTokensToSample;
  public int topK;
  public int topP;

  public CompletionsInput(
      @NotNull List<Message> messages,
      float temperature,
      List<String> stopSequences,
      int maxTokensToSample,
      int topK,
      int topP) {
    this.messages = messages;
    this.temperature = temperature;
    this.stopSequences = stopSequences;
    this.maxTokensToSample = maxTokensToSample;
    this.topK = topK;
    this.topP = topP;
  }
}
