package com.sourcegraph.cody.autocomplete;

import org.jetbrains.annotations.NotNull;

/** A class that stores the state and timing information of an autocompletion. */
public class AutocompleteTelemetry {
  private long completionTriggeredTimestampMs;
  private long completionDisplayedTimestampMs;
  private long completionHiddenTimestampMs;

  public static @NotNull AutocompleteTelemetry createAndMarkTriggered() {
    var autocompletion = new AutocompleteTelemetry();
    // TODO: we could use java.time.Instant.now() and java.time.Duration.between(Instant,Instant)
    // and avoid the "TimestampMs" suffixes
    autocompletion.completionTriggeredTimestampMs = System.currentTimeMillis();
    return autocompletion;
  }

  private AutocompleteTelemetry() {}

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
