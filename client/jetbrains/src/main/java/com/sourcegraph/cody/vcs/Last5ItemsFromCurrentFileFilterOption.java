package com.sourcegraph.cody.vcs;

import com.intellij.openapi.vcs.FilePath;
import java.util.function.Supplier;
import org.jetbrains.annotations.NotNull;

public class Last5ItemsFromCurrentFileFilterOption implements VcsLogFilterOption {

  private final @NotNull FilePath filePath;
  private final @NotNull String label;

  private final @NotNull String filterDescription;

  public Last5ItemsFromCurrentFileFilterOption(@NotNull FilePath filePath) {
    this.filePath = filePath;
    this.label = "Last 5 items for " + filePath.getName();
    this.filterDescription = "What changed in " + filePath.getName() + " in the last 5 commits?";
  }

  @Override
  public @NotNull String label() {
    return label;
  }

  @Override
  public @NotNull Supplier<VcsFilter> getFilterSupplier() {
    return () -> VcsFilter.last5ItemsForFile(filePath, filterDescription);
  }
}
