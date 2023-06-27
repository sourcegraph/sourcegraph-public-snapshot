package com.sourcegraph.common;

import org.jetbrains.annotations.NotNull;

public class AuthorizationUtil {
  public static boolean isValidAccessToken(@NotNull String accessToken) {
    return accessToken.isEmpty()
        || accessToken.length() == 40
        || (accessToken.startsWith("sgp_") && accessToken.length() == 44);
  }
}
