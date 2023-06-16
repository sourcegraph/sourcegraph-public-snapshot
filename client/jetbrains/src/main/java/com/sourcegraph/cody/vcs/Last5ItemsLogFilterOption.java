package com.sourcegraph.cody.vcs;

import java.util.function.Supplier;
import org.jetbrains.annotations.NotNull;

public class Last5ItemsLogFilterOption implements VcsLogFilterOption {

  @Override
  public @NotNull String label() {
    return "Last 5 items";
  }

  @Override
  public @NotNull Supplier<VcsFilter> getFilterSupplier() {
    return VcsFilter::last5Items;
  }
}
