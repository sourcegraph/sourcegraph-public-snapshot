package com.sourcegraph.cody.autocomplete;

import com.sourcegraph.cody.agent.protocol.CompletionEvent;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

/** A class that stores the state and timing information of an autocompletion. */
public class AutocompleteTelemetry {
  private long completionTriggeredTimestampMs;
  private long completionDisplayedTimestampMs;
  private long completionHiddenTimestampMs;
  @Nullable private CompletionEvent completionEvent;

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

  public void markCompletionEvent(@Nullable CompletionEvent event) {
    this.completionEvent = event;
  }

  public void markCompletionHidden() {
    this.completionHiddenTimestampMs = System.currentTimeMillis();
  }

  public long getDisplayDurationMs() {
    return completionHiddenTimestampMs - completionDisplayedTimestampMs;
  }

  @Nullable
  public CompletionEvent.ContextSummary contextSummary() {
    return (completionEvent != null && completionEvent.params != null)
        ? completionEvent.params.contextSummary
        : null;
  }

  @Nullable
  public CompletionEvent.Params params() {
    return (completionEvent != null) ? completionEvent.params : null;
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
