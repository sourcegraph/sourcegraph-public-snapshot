package com.sourcegraph.cody.chat;

import static com.sourcegraph.cody.TruncationUtils.CHARS_PER_TOKEN;

import com.sourcegraph.cody.api.Message;
import com.sourcegraph.cody.context.ContextMessage;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

public class Transcript {

  private final List<Interaction> interactions = new ArrayList<>();

  public void addInteraction(Interaction interaction) {
    interactions.add(interaction);
  }

  public Optional<Interaction> getLastInteraction() {
    if (this.isEmpty()) {
      return Optional.empty();
    }
    return Optional.ofNullable(interactions.get(interactions.size() - 1));
  }

  public boolean isEmpty() {
    return interactions.isEmpty();
  }

  public void addAssistantResponse(ChatMessage assistantMessage) {
    this.getLastInteraction()
        .ifPresent(interaction -> interaction.setAssistantMessage(assistantMessage));
  }

  private int getLastInteractionWithContextIndex() {
    for (int index = this.interactions.size() - 1; index >= 0; index--) {
      boolean hasContext = this.interactions.get(index).hasContext();
      if (hasContext) {
        return index;
      }
    }
    return -1;
  }

  public List<Message> getPromptForLastInteraction(List<Message> preamble, int maxPromptLength) {
    if (this.isEmpty()) {
      return Collections.emptyList();
    }

    int lastInteractionWithContextIndex = getLastInteractionWithContextIndex();
    List<Message> messages = new ArrayList<>();
    for (int index = 0; index < this.interactions.size(); index++) {
      Interaction interaction = this.interactions.get(index);
      Message humanMessage = interaction.getHumanMessage();
      Optional<ChatMessage> assistantMessage = interaction.getAssistantMessage();
      List<ContextMessage> contextMessages = interaction.getFullContext();
      if (index == lastInteractionWithContextIndex) {
        messages.addAll(contextMessages);
      }
      messages.add(humanMessage);
      assistantMessage.ifPresent(messages::add);
    }

    int preambleTokensUsage = 0;
    for (Message message : preamble) {
      preambleTokensUsage += estimateTokensUsage(message);
    }
    List<Message> truncatedMessages =
        truncatePrompt(messages, maxPromptLength - preambleTokensUsage);

    // Filter out extraneous fields from ContextMessage instances
    truncatedMessages =
        truncatedMessages.stream()
            .map(m -> new Message(m.getSpeaker(), m.getText()))
            .collect(Collectors.toList());

    List<Message> allMessages = new ArrayList<>();
    allMessages.addAll(preamble);
    allMessages.addAll(truncatedMessages);
    return allMessages;
  }

  private int estimateTokensUsage(Message message) {
    return message.getText().length() / CHARS_PER_TOKEN;
  }

  /**
   * Truncates the given prompt messages to fit within the available tokens budget. The truncation
   * is done by removing the oldest pairs of messages first. No individual message will be
   * truncated. We just remove pairs of messages if they exceed the available tokens budget.
   */
  private List<Message> truncatePrompt(List<Message> messages, int maxTokens) {
    List<Message> newPromptMessages = new ArrayList<>();
    int availablePromptTokensBudget = maxTokens;
    for (int i = messages.size() - 1; i >= 1; i -= 2) {
      Message humanMessage = messages.get(i - 1);
      Message botMessage = messages.get(i);
      int combinedTokensUsage = estimateTokensUsage(humanMessage) + estimateTokensUsage(botMessage);

      // We stop adding pairs of messages once we exceed the available tokens budget.
      if (combinedTokensUsage <= availablePromptTokensBudget) {
        newPromptMessages.add(botMessage);
        newPromptMessages.add(humanMessage);
        availablePromptTokensBudget -= combinedTokensUsage;
      } else {
        break;
      }
    }

    // Reverse the prompt messages, so they appear in chat order (older -> newer).
    Collections.reverse(newPromptMessages);
    return newPromptMessages;
  }
}
