package com.sourcegraph.config;

import com.intellij.openapi.components.PersistentStateComponent;
import com.intellij.openapi.components.State;
import com.intellij.openapi.components.Storage;
import com.sourcegraph.find.Search;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

@State(
    name = "Config",
    storages = {@Storage("sourcegraph.xml")})
public class SourcegraphProjectService
    implements PersistentStateComponent<SourcegraphProjectService>, SourcegraphService {
  @Nullable public String instanceType;
  @Nullable public String url;
  @Nullable public String dotComAccessToken;
  @Nullable public String enterpriseAccessToken;
  @Nullable public String customRequestHeaders;
  @Nullable public String defaultBranch;
  @Nullable public String remoteUrlReplacements;
  @Nullable public String lastSearchQuery;
  public boolean lastSearchCaseSensitive;
  @Nullable public String lastSearchPatternType;
  @Nullable public String lastSearchContextSpec;

  @Override
  @Nullable
  public String getInstanceType() {
    return instanceType;
  }

  @Override
  @Nullable
  public String getSourcegraphUrl() {
    return url;
  }

  @Override
  @Nullable
  public String getDotComAccessToken() {
    if (dotComAccessToken == null) {
      return System.getenv("SRC_ACCESS_TOKEN");
    }
    return dotComAccessToken;
  }

  @Override
  @Nullable
  public String getCustomRequestHeaders() {
    return customRequestHeaders;
  }

  @Override
  @Nullable
  public String getDefaultBranchName() {
    return defaultBranch;
  }

  @Override
  @Nullable
  public String getRemoteUrlReplacements() {
    return remoteUrlReplacements;
  }

  @Override
  @Nullable
  public Search getLastSearch() {
    if (lastSearchQuery == null) {
      return null;
    } else {
      return new Search(
          lastSearchQuery, lastSearchCaseSensitive, lastSearchPatternType, lastSearchContextSpec);
    }
  }

  @Nullable
  @Override
  public SourcegraphProjectService getState() {
    return this;
  }

  @Override
  public void loadState(@NotNull SourcegraphProjectService settings) {
    this.instanceType = settings.instanceType;
    this.url = settings.url;
    this.dotComAccessToken = settings.dotComAccessToken;
    this.enterpriseAccessToken = settings.enterpriseAccessToken;
    this.customRequestHeaders = settings.customRequestHeaders;
    this.defaultBranch = settings.defaultBranch;
    this.remoteUrlReplacements = settings.remoteUrlReplacements;
    this.lastSearchQuery = settings.lastSearchQuery != null ? settings.lastSearchQuery : "";
    this.lastSearchCaseSensitive = settings.lastSearchCaseSensitive;
    this.lastSearchPatternType =
        settings.lastSearchPatternType != null ? settings.lastSearchPatternType : "literal";
    this.lastSearchContextSpec =
        settings.lastSearchContextSpec != null ? settings.lastSearchContextSpec : "global";
  }

  @Override
  public String getEnterpriseAccessToken() {
    return enterpriseAccessToken == null
        ? System.getenv("SRC_ACCESS_TOKEN")
        : enterpriseAccessToken;
  }

  @Override
  public boolean areChatPredictionsEnabled() {
    // TODO: implement
    return false;
  }

  @Override
  public String getCodebase() {
    // TODO: implement
    return null;
  }
}
