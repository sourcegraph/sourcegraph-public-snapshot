package com.sourcegraph.cody.autocomplete.prompt_library;

import com.sourcegraph.cody.api.Message;
import java.util.List;

public class CompletionParameters {

  public List<Message> messages;
  public int maxTokensToSample;
  public float temperature;
  public List<String> stopSequences;
  public int topK;
  public int topP;
  public String model;

  // builder methods
  public CompletionParameters withMessages(List<Message> messages) {
    this.messages = messages;
    return this;
  }

  public CompletionParameters withMaxTokensToSample(int maxTokensToSample) {
    this.maxTokensToSample = maxTokensToSample;
    return this;
  }

  public CompletionParameters withTemperature(float temperature) {
    this.temperature = temperature;
    return this;
  }

  public CompletionParameters withStopSequences(List<String> stopSequences) {
    this.stopSequences = stopSequences;
    return this;
  }

  public CompletionParameters withTopK(int topK) {
    this.topK = topK;
    return this;
  }

  public CompletionParameters withTopP(int topP) {
    this.topP = topP;
    return this;
  }

  public CompletionParameters withModel(String model) {
    this.model = model;
    return this;
  }
}
