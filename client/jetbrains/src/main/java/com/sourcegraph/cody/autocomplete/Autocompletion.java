package com.sourcegraph.cody.autocomplete;

import org.jetbrains.annotations.NotNull;

public class Autocompletion {
  private long completionTriggeredTimestampMs;
  private long completionDisplayedTimestampMs;
  private long completionHiddenTimestampMs;

  public static @NotNull Autocompletion createAndMarkTriggered() {
    var autocompletion = new Autocompletion();
    autocompletion.completionTriggeredTimestampMs = System.currentTimeMillis();
    return autocompletion;
  }

  private Autocompletion() {}

  public void markCompletionDisplayed() {
    this.completionDisplayedTimestampMs = System.currentTimeMillis();
  }

  public long getLatencyMs() {
    return completionDisplayedTimestampMs - completionTriggeredTimestampMs;
  }

  public void markCompletionHidden() {
    this.completionHiddenTimestampMs = System.currentTimeMillis();
  }

  public long getDisplayDurationMs() {
    return completionHiddenTimestampMs - completionDisplayedTimestampMs;
  }

  public @NotNull AutocompletionStatus getStatus() {
    if (completionDisplayedTimestampMs == 0) {
      return AutocompletionStatus.TRIGGERED_NOT_DISPLAYED;
    }
    if (completionHiddenTimestampMs == 0) {
      return AutocompletionStatus.DISPLAYED;
    }
    return AutocompletionStatus.HIDDEN;
  }
}
