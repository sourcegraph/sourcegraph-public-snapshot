package com.sourcegraph.cody.autocomplete;

import java.util.Arrays;
import java.util.Optional;
import org.jetbrains.annotations.NotNull;

public enum AutoCompleteProviderType {
  ANTHROPIC,
  UNSTABLE_CODEGEN,
  UNSTABLE_AZURE_OPENAI;

  public static final AutoCompleteProviderType DEFAULT_AUTOCOMPLETE_PROVIDER_TYPE = ANTHROPIC;

  @NotNull
  public static Optional<AutoCompleteProviderType> optionalValueOf(@NotNull String name) {
    return Arrays.stream(AutoCompleteProviderType.values())
        .filter(providerType -> providerType.vscodeSettingString().equals(name))
        .findFirst();
  }

  public String vscodeSettingString() {
    return super.toString().toLowerCase().replace('_', '-');
  }
}
