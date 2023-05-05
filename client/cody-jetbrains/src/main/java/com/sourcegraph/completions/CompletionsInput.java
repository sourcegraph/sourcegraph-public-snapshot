package com.sourcegraph.completions;

import java.util.List;

/**
 * Input for the completions request.
 */
public class CompletionsInput {
    public List<Message> messages;
    public float temperature;
    public int maxTokensToSample;
    public int topK;
    public int topP;

    public CompletionsInput(List<Message> messages, float temperature, int maxTokensToSample, int topK, int topP) {
        this.messages = messages;
        this.temperature = temperature;
        this.maxTokensToSample = maxTokensToSample;
        this.topK = topK;
        this.topP = topP;
    }

    public void addMessage(Speaker speaker, String text) {
        messages.add(new Message(speaker, text));
    }
}
