package com.sourcegraph.config;

import com.intellij.credentialStore.CredentialAttributes;
import com.intellij.credentialStore.CredentialAttributesKt;
import com.intellij.credentialStore.Credentials;
import com.intellij.ide.passwordSafe.PasswordSafe;
import java.util.List;
import java.util.Optional;
import org.apache.commons.lang3.StringUtils;
import org.jetbrains.annotations.NotNull;

public class AccessTokenStorage {

  private static final String SOURCEGRAPH = "sourcegraph";
  private static final String ACCESS_TOKEN_KEY = "accessToken";
  private static final String ENTERPRISE = "enterprise";
  private static final String DOT_COM = "dotcom";
  private static final List<String> enterpriseAccessTokenKeyParts =
      List.of(ACCESS_TOKEN_KEY, ENTERPRISE);

  private static final List<String> dotComAccessTokenKeyParts = List.of(ACCESS_TOKEN_KEY, DOT_COM);

  @NotNull
  public static Optional<String> getEnterpriseAccessToken() {
    return getApplicationLevelSecureConfig(enterpriseAccessTokenKeyParts);
  }

  public static void setApplicationEnterpriseAccessToken(@NotNull String accessToken) {
    setApplicationLevelSecureConfig(enterpriseAccessTokenKeyParts, accessToken);
  }

  @NotNull
  public static Optional<String> getDotComAccessToken() {
    return getApplicationLevelSecureConfig(dotComAccessTokenKeyParts);
  }

  public static void setApplicationDotComAccessToken(@NotNull String accessToken) {
    setApplicationLevelSecureConfig(dotComAccessTokenKeyParts, accessToken);
  }

  @NotNull
  private static Optional<String> getApplicationLevelSecureConfig(@NotNull List<String> keyParts) {
    return Optional.ofNullable(
        PasswordSafe.getInstance().getPassword(createCredentialAttributes(createKey(keyParts))));
  }

  private static void setApplicationLevelSecureConfig(
      @NotNull List<String> keyParts, @NotNull String accessToken) {
    CredentialAttributes credentialAttributes = createCredentialAttributes(createKey(keyParts));
    Credentials credentials = new Credentials(null, accessToken);
    PasswordSafe.getInstance().set(credentialAttributes, credentials);
  }

  @NotNull
  private static String createKey(@NotNull List<String> keyParts) {
    return StringUtils.join(keyParts, ".");
  }

  @NotNull
  private static CredentialAttributes createCredentialAttributes(@NotNull String key) {
    return new CredentialAttributes(CredentialAttributesKt.generateServiceName(SOURCEGRAPH, key));
  }
}
