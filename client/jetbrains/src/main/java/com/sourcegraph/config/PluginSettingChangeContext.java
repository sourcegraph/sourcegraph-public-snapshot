package com.sourcegraph.config;

import org.jetbrains.annotations.Nullable;

public class PluginSettingChangeContext {
    @Nullable
    public final String oldUrl;

    @Nullable
    public final String oldAccessToken;

    @Nullable
    public final String newUrl;

    @Nullable
    public final String newAccessToken;

    @Nullable
    public final String newCustomRequestHeaders;

    public PluginSettingChangeContext(@Nullable String oldUrl,
                                      @Nullable String oldAccessToken,
                                      @Nullable String newUrl,
                                      @Nullable String newAccessToken,
                                      @Nullable String newCustomRequestHeaders) {
        this.oldUrl = oldUrl;
        this.oldAccessToken = oldAccessToken;
        this.newUrl = newUrl;
        this.newAccessToken = newAccessToken;
        this.newCustomRequestHeaders = newCustomRequestHeaders;
    }
}
