package com.sourcegraph.cody.localapp;

import com.sourcegraph.common.AuthorizationUtil;
import java.net.ConnectException;
import java.nio.file.Path;
import java.util.Map;
import java.util.Optional;
import org.apache.commons.lang.SystemUtils;
import org.apache.http.HttpResponse;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClientBuilder;
import org.apache.http.util.EntityUtils;
import org.jetbrains.annotations.NotNull;

public class LocalAppManager {
  public static final String DEFAULT_LOCAL_APP_URL = "http://localhost:3080/";
  private static final Map<String, LocalAppPaths> appPathsByPlatform =
      Map.of(
          "darwin", // only support macOS for now
          new LocalAppPaths(
              Path.of("/Applications/Sourcegraph.app"),
              Path.of("/Applications/Cody.app"),
              Path.of(
                  SystemUtils.getUserHome()
                      + "/Library/Application Support/com.sourcegraph.cody/site.config.json"),
              Path.of(
                  SystemUtils.getUserHome()
                      + "/Library/Application Support/com.sourcegraph.cody/app.json")));

  public static boolean isLocalAppInstalled() {
    return getLocalAppPaths().map(LocalAppPaths::anyPathExists).orElse(false);
  }

  @NotNull
  public static Optional<LocalAppPaths> getLocalAppPaths() {
    return Optional.ofNullable(System.getProperty("os.name"))
        .map(String::toLowerCase)
        .map(
            osName -> {
              if (osName.contains("mac")) return "darwin";
              else if (osName.contains("windows")) return "windows";
              else return osName;
            })
        .flatMap(osKey -> Optional.ofNullable(appPathsByPlatform.get(osKey)));
  }

  public static boolean isPlatformSupported() {
    return getLocalAppPaths().isPresent();
  }

  @NotNull
  public static Optional<String> getLocalAppAccessToken() {
    return getLocalAppInfo()
        .flatMap(appInfo -> Optional.ofNullable(appInfo.getToken()))
        .filter(AuthorizationUtil::isValidAccessToken);
  }

  @NotNull
  public static Optional<LocalAppInfo> getLocalAppInfo() {
    return getLocalAppPaths().flatMap(LocalAppPaths::getAppInfo);
  }

  public static boolean isLocalAppRunning() {
    return getRunningAppVersion().isPresent();
  }

  /**
   * @return gets the local app url from the local app json file, falls back to
   *     [[DEFAULT_LOCAL_APP_URL]] if not present
   */
  @NotNull
  public static String getLocalAppUrl() {
    return getLocalAppInfo()
        .flatMap(appInfo -> Optional.ofNullable(appInfo.getEndpoint()))
        .map(endpoint -> endpoint.endsWith("/") ? endpoint : endpoint + "/")
        .orElse(DEFAULT_LOCAL_APP_URL);
  }

  @NotNull
  private static Optional<String> getRunningAppVersion() {
    // TODO: do this asynchronously
    try (CloseableHttpClient httpClient = HttpClientBuilder.create().build()) {
      HttpGet request = new HttpGet(getLocalAppUrl() + "/__version");
      HttpResponse response = httpClient.execute(request);
      int statusCode = response.getStatusLine().getStatusCode();
      String responseBody = EntityUtils.toString(response.getEntity());
      if (statusCode != 200) {
        System.err.println(
            "Could not fetch local Cody app version. Got status code "
                + statusCode
                + ": "
                + responseBody);
        return Optional.empty();
      } else {
        System.out.println("Running local Cody app version: " + responseBody);
        return Optional.of(responseBody);
      }
    } catch (ConnectException e) {
      System.err.println("Could not connect to the local Cody app.");
      return Optional.empty();
    } catch (Exception e) {
      e.printStackTrace();
      return Optional.empty();
    }
  }
}
