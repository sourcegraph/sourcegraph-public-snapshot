package com.sourcegraph.cody.vcs;

import static java.time.temporal.ChronoUnit.DAYS;

import java.time.Instant;
import java.util.Date;
import java.util.function.Supplier;
import org.jetbrains.annotations.NotNull;

public class LastNDaysFilterOption implements VcsLogFilterOption {

  private final int numberOfDays;
  private final @NotNull String label;
  private final @NotNull String filterDescription;

  public LastNDaysFilterOption(int days, @NotNull String label, @NotNull String filterDescription) {
    this.numberOfDays = days;
    this.label = label;
    this.filterDescription = filterDescription;
  }

  @Override
  public @NotNull String label() {
    return label;
  }

  @Override
  public @NotNull Supplier<VcsFilter> getFilterSupplier() {
    return () ->
        VcsFilter.fromDate(Date.from(Instant.now().minus(numberOfDays, DAYS)), filterDescription);
  }
}
