package com.sourcegraph.cody.vcs;

import java.util.function.Supplier;
import org.jetbrains.annotations.NotNull;

public interface VcsLogFilterOption {
  @NotNull
  String label();

  @NotNull
  Supplier<VcsFilter> getFilterSupplier();
}
