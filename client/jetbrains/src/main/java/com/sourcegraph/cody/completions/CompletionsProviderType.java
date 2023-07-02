package com.sourcegraph.cody.completions;

import java.util.Optional;
import org.jetbrains.annotations.NotNull;

public enum CompletionsProviderType {
  ANTHROPIC,
  UNSTABLE_CODEGEN;

  public static final CompletionsProviderType DEFAULT_COMPLETIONS_PROVIDER_TYPE = ANTHROPIC;

  @NotNull
  public static Optional<CompletionsProviderType> optionalValueOf(@NotNull String name) {
    String normalizedName = name.trim().toUpperCase().replace('-', '_');
    try {
      return Optional.of(CompletionsProviderType.valueOf(normalizedName));
    } catch (IllegalArgumentException e) {
      System.err.println("Cody: Error: Invalid completions provider type: " + name);
      return Optional.empty();
    }
  }
}
