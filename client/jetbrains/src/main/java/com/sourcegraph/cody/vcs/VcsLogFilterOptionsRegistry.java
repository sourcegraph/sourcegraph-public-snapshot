package com.sourcegraph.cody.vcs;

import java.util.*;
import java.util.function.Supplier;
import org.jetbrains.annotations.NotNull;

public class VcsLogFilterOptionsRegistry {

  private final @NotNull Map<String, Supplier<VcsFilter>> allOptions = new HashMap<>();
  private final @NotNull List<String> optionLabels = new ArrayList<>();

  public VcsLogFilterOptionsRegistry() {
    List.of(
            new Last5ItemsLogFilterOption(),
            new LastNDaysFilterOption(
                1, "Last day", "What has changed in my codebase in the last day?"),
            new LastNDaysFilterOption(
                7, "Last week", "What changed in my codebase in the last week?"))
        .forEach(
            vcsLogFilterOption -> {
              allOptions.put(vcsLogFilterOption.label(), vcsLogFilterOption.getFilterSupplier());
              optionLabels.add(vcsLogFilterOption.label());
            });
  }

  public @NotNull List<String> getAllOptions() {
    return optionLabels;
  }

  public @NotNull Supplier<VcsFilter> getFilterSupplierForOption(String option) {
    return allOptions.getOrDefault(option, VcsFilter::emptyFilters);
  }

  public void addFilterOption(VcsLogFilterOption vcsLogFilterOption) {
    allOptions.put(vcsLogFilterOption.label(), vcsLogFilterOption.getFilterSupplier());
    optionLabels.add(vcsLogFilterOption.label());
  }
}
