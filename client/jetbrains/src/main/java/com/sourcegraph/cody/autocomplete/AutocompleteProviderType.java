package com.sourcegraph.cody.autocomplete;

import java.util.Arrays;
import java.util.Optional;
import org.jetbrains.annotations.NotNull;

public enum AutocompleteProviderType {
  ANTHROPIC,
  UNSTABLE_CODEGEN,
  UNSTABLE_OPENAI,
  UNSTABLE_FIREWORKS;

  public static Optional<AutocompleteProviderType> optionalValueOf(@NotNull String name) {
    return Arrays.stream(AutocompleteProviderType.values())
        .filter(providerType -> providerType.vscodeSettingString().equals(name))
        .findFirst();
  }

  public String vscodeSettingString() {
    return super.toString().toLowerCase().replace('_', '-');
  }
}
