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
    public final String newRequestHeaders;

    public PluginSettingChangeContext(@Nullable String oldUrl,
                                      @Nullable String oldAccessToken,
                                      @Nullable String newUrl,
                                      @Nullable String newAccessToken,
                                      @Nullable String newRequestHeaders) {
        this.oldUrl = oldUrl;
        this.oldAccessToken = oldAccessToken;
        this.newUrl = newUrl;
        this.newAccessToken = newAccessToken;
        this.newRequestHeaders = newRequestHeaders;
    }
}
