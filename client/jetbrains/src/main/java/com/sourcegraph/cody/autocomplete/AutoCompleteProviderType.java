package com.sourcegraph.cody.autocomplete;

import com.intellij.openapi.diagnostic.Logger;
import java.util.Optional;
import org.jetbrains.annotations.NotNull;

public enum AutoCompleteProviderType {
  ANTHROPIC,
  UNSTABLE_CODEGEN;

  public static final AutoCompleteProviderType DEFAULT_AUTOCOMPLETE_PROVIDER_TYPE = ANTHROPIC;
  private static final Logger logger = Logger.getInstance(AutoCompleteProviderType.class);

  @NotNull
  public static Optional<AutoCompleteProviderType> optionalValueOf(@NotNull String name) {
    String normalizedName = name.trim().toUpperCase().replace('-', '_');
    try {
      return Optional.of(AutoCompleteProviderType.valueOf(normalizedName));
    } catch (IllegalArgumentException e) {
      logger.warn("Cody: Error: Invalid autocomplete provider type: " + name);
      return Optional.empty();
    }
  }
}
