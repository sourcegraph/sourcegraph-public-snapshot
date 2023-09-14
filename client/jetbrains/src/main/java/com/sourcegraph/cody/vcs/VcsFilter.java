package com.sourcegraph.cody.vcs;

import com.intellij.openapi.vcs.FilePath;
import com.intellij.vcs.log.VcsLogFilterCollection;
import com.intellij.vcs.log.visible.filters.VcsLogFilterObject;
import java.util.Collections;
import java.util.Date;
import org.jetbrains.annotations.NotNull;

public class VcsFilter {

  private static final int UNLIMITED_RESULTS = -1;
  private final int count;
  private final @NotNull VcsLogFilterCollection filterCollection;
  private final @NotNull String filterDescription;

  private VcsFilter(
      int count,
      @NotNull VcsLogFilterCollection filterCollection,
      @NotNull String filterDescription) {
    this.count = count;
    this.filterCollection = filterCollection;
    this.filterDescription = filterDescription;
  }

  public static VcsFilter last5Items() {
    return new VcsFilter(
        5,
        VcsLogFilterObject.EMPTY_COLLECTION,
        "What changed in my codebase in the last 5 commits?");
  }

  public static VcsFilter fromDate(Date after, String filterDescription) {
    return new VcsFilter(
        UNLIMITED_RESULTS,
        VcsLogFilterObject.collection(VcsLogFilterObject.fromDates(after, null)),
        filterDescription);
  }

  public static VcsFilter last5ItemsForFile(FilePath filePath, String filterDescription) {
    return new VcsFilter(
        5,
        VcsLogFilterObject.collection(
            VcsLogFilterObject.fromPaths(Collections.singleton(filePath))),
        filterDescription);
  }

  public static VcsFilter emptyFilters() {
    return new VcsFilter(
        UNLIMITED_RESULTS, VcsLogFilterObject.EMPTY_COLLECTION, "What changed in my codebase?");
  }

  public int getCount() {
    return count;
  }

  public @NotNull VcsLogFilterCollection getFilterCollection() {
    return filterCollection;
  }

  public @NotNull String getFilterDescription() {
    return filterDescription;
  }
}
