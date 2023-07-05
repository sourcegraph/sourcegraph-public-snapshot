package com.sourcegraph.cody.localapp;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.intellij.openapi.diagnostic.Logger;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Optional;

public class LocalAppPaths {
  private static final Logger logger = Logger.getInstance(LocalAppPaths.class);
  public final Path sourcegraphAppFile;
  public final Path codyAppFile;
  public final Path siteConfigJsonFile;
  public final Path appJsonFile;

  LocalAppPaths(
      Path sourcegraphAppFile, Path codyAppFile, Path siteConfigJsonFile, Path appJsonFile) {
    this.sourcegraphAppFile = sourcegraphAppFile;
    this.codyAppFile = codyAppFile;
    this.siteConfigJsonFile = siteConfigJsonFile;
    this.appJsonFile = appJsonFile;
  }

  public boolean anyPathExists() {
    return pathExists(sourcegraphAppFile)
        || pathExists(codyAppFile)
        || pathExists(siteConfigJsonFile)
        || pathExists(appJsonFile);
  }

  public boolean isCodyInstalled() {
    return pathExists(codyAppFile);
  }

  private static boolean pathExists(Path path) {
    try {
      return path.toFile().exists();
    } catch (SecurityException | UnsupportedOperationException e) {
      // in case we can't read the path for whatever reason, we assume it doesn't exist
      return false;
    }
  }

  public Optional<LocalAppInfo> getAppInfo() {
    if (pathExists(appJsonFile)) {
      try {
        String jsonContent = Files.readString(appJsonFile);

        ObjectMapper objectMapper = new ObjectMapper();
        LocalAppInfo appInfo = objectMapper.readValue(jsonContent, LocalAppInfo.class);
        return Optional.of(appInfo);
      } catch (IOException e) {
        logger.error("Error reading local Cody app JSON file: " + e.getMessage());
        return Optional.empty();
      }
    } else return Optional.empty();
  }
}
