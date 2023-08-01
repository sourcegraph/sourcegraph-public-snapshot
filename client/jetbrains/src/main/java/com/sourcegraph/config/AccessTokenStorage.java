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
  private static String cachedEnterpriseAccessToken = null;
  private static String cachedDotComAccessToken = null;

  private static final List<String> enterpriseAccessTokenKeyParts =
      List.of("accessToken", "enterprise");

  private static final List<String> dotComAccessTokenKeyParts = List.of("accessToken", "dotcom");

  // Empty string means empty or no token. No value means user denied access to token.
  @NotNull
  public static Optional<String> getEnterpriseAccessToken() {
    if (cachedEnterpriseAccessToken != null) {
      return Optional.of(cachedEnterpriseAccessToken);
    }
    Optional<String> token = getApplicationLevelSecureConfig(enterpriseAccessTokenKeyParts);
    token.ifPresent(t -> cachedEnterpriseAccessToken = t);
    return token;
  }

  public static void setApplicationEnterpriseAccessToken(@NotNull String accessToken) {
    setApplicationLevelSecureConfig(enterpriseAccessTokenKeyParts, accessToken);
    cachedEnterpriseAccessToken = accessToken;
  }

  // Empty string means empty or no token. No value means user denied access to token.
  @NotNull
  public static Optional<String> getDotComAccessToken() {
    if (cachedDotComAccessToken != null) {
      return Optional.of(cachedDotComAccessToken);
    }
    Optional<String> token = getApplicationLevelSecureConfig(dotComAccessTokenKeyParts);
    token.ifPresent(t -> cachedDotComAccessToken = t);
    return token;
  }

  public static void setApplicationDotComAccessToken(@NotNull String accessToken) {
    setApplicationLevelSecureConfig(dotComAccessTokenKeyParts, accessToken);
    cachedDotComAccessToken = accessToken;
  }

  // Empty string means empty or no password. No value means user denied access to password.
  @NotNull
  private static Optional<String> getApplicationLevelSecureConfig(@NotNull List<String> keyParts) {
    CredentialAttributes credentialAttributes = createCredentialAttributes(createKey(keyParts));
    Credentials credentials = PasswordSafe.getInstance().get(credentialAttributes);
    // No credentials found
    if (credentials == null) {
      return Optional.of("");
    }
    String password = credentials.getPasswordAsString();
    if (password == null) { // User denied access to password
      return Optional.empty();
    }
    return Optional.of(password);
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
    return new CredentialAttributes(CredentialAttributesKt.generateServiceName("sourcegraph", key));
  }
}
