package com.sourcegraph.cody.completions;

import com.intellij.openapi.diagnostic.Logger;
import java.util.Optional;
import org.jetbrains.annotations.NotNull;

public enum CompletionsProviderType {
  ANTHROPIC,
  UNSTABLE_CODEGEN;

  public static final CompletionsProviderType DEFAULT_COMPLETIONS_PROVIDER_TYPE = ANTHROPIC;
  private static final Logger logger = Logger.getInstance(CompletionsProviderType.class);

  @NotNull
  public static Optional<CompletionsProviderType> optionalValueOf(@NotNull String name) {
    String normalizedName = name.trim().toUpperCase().replace('-', '_');
    try {
      return Optional.of(CompletionsProviderType.valueOf(normalizedName));
    } catch (IllegalArgumentException e) {
      logger.error("Cody: Error: Invalid completions provider type: " + name);
      return Optional.empty();
    }
  }
}
