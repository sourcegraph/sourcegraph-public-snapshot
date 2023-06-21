package com.sourcegraph.cody.vscode;

import java.util.Objects;

public class Range {
  public final Position start;
  public final Position end;

  public Range(Position start, Position end) {
    this.start = start;
    this.end = end;
  }

  public Range withStart(Position newStart) {
    return new Range(newStart, this.end);
  }

  public Range withEnd(Position newEnd) {
    return new Range(this.start, newEnd);
  }

  @Override
  public String toString() {
    return "Range{" + "start=" + start + ", end=" + end + '}';
  }

  @Override
  public boolean equals(Object o) {
    if (this == o) return true;
    if (!(o instanceof Range)) return false;
    Range range = (Range) o;
    return Objects.equals(start, range.start) && Objects.equals(end, range.end);
  }

  @Override
  public int hashCode() {
    return Objects.hash(start, end);
  }
}
