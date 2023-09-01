package com.sourcegraph.cody.agent.protocol;

import org.jetbrains.annotations.Nullable;

public class CompletionEvent {

  public static class Params {
    public String type;
    public boolean multiline;
    public String multilineMode;
    public String providerIdentifier;
    public String languageId;
    @Nullable public ContextSummary contextSummary;
    @Nullable public String source;
    public String id;
    @Nullable public Integer lineCount;
    @Nullable public Integer charCount;

    @Override
    public String toString() {
      return "Params{"
          + "type='"
          + type
          + '\''
          + ", multiline="
          + multiline
          + ", multilineMode='"
          + multilineMode
          + '\''
          + ", providerIdentifier='"
          + providerIdentifier
          + '\''
          + ", languageId='"
          + languageId
          + '\''
          + ", contextSummary="
          + contextSummary
          + ", source='"
          + source
          + '\''
          + ", id='"
          + id
          + '\''
          + ", lineCount="
          + lineCount
          + ", charCount="
          + charCount
          + '}';
    }
  }

  public static class ContextSummary {
    @Nullable public Double embeddings;
    @Nullable public Double local;
    public double duration;

    @Override
    public String toString() {
      return "ContextSummary{"
          + "embeddings="
          + embeddings
          + ", local="
          + local
          + ", duration="
          + duration
          + '}';
    }
  }

  @Nullable public Params params;
  public double startedAt;
  public double networkRequestStartedAt;
  public double startLoggedAt;
  public double loadedAt;
  public double suggestedAt;
  public double suggestionLoggedAt;
  public double acceptedAt;

  @Override
  public String toString() {
    return "CompletionEvent{"
        + "params="
        + params
        + ", startedAt="
        + startedAt
        + ", networkRequestStartedAt="
        + networkRequestStartedAt
        + ", startLoggedAt="
        + startLoggedAt
        + ", loadedAt="
        + loadedAt
        + ", suggestedAt="
        + suggestedAt
        + ", suggestionLoggedAt="
        + suggestionLoggedAt
        + ", acceptedAt="
        + acceptedAt
        + '}';
  }
}
