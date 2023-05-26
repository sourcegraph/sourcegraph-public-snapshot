package com.sourcegraph.cody.config;

import org.jetbrains.annotations.Nullable;

public interface CodyService {
  @Nullable
  String getInstanceType();

  void setInstanceType(@Nullable String instanceType);

  @Nullable
  String getDotcomAccessToken();

  void setDotcomAccessToken(@Nullable String dotcomAccessToken);

  @Nullable
  String getEnterpriseUrl();

  void setEnterpriseUrl(@Nullable String enterpriseUrl);

  @Nullable
  String getEnterpriseAccessToken();

  void setEnterpriseAccessToken(@Nullable String enterpriseAccessToken);

  @Nullable
  String getCustomRequestHeaders();

  void setCustomRequestHeaders(@Nullable String customRequestHeaders);

  @Nullable
  String getCodebase();

  void setCodebase(@Nullable String codebase);

  @Nullable
  Boolean areChatPredictionsEnabled();

  void setChatPredictionsEnabled(@Nullable Boolean areChatPredictionsEnabled);
}
