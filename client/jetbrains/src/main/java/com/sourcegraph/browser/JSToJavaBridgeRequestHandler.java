package com.sourcegraph.browser;

import com.google.gson.JsonObject;
import com.intellij.openapi.project.Project;
import com.intellij.ui.jcef.JBCefJSQuery;
import com.sourcegraph.config.ConfigUtil;
import com.sourcegraph.config.ThemeUtil;
import com.sourcegraph.find.PreviewContent;
import com.sourcegraph.find.PreviewPanel;
import com.sourcegraph.find.Search;
import org.jetbrains.annotations.NotNull;

import javax.annotation.Nullable;
import java.io.PrintWriter;
import java.io.StringWriter;

public class JSToJavaBridgeRequestHandler {
    private final Project project;
    private final PreviewPanel previewPanel;
    private final BrowserAndLoadingPanel topPanel;

    public JBCefJSQuery.Response handle(@NotNull JsonObject request) {
        String action = request.get("action").getAsString();
        JsonObject arguments;
        PreviewContent previewContent;
        try {
            switch (action) {
                case "getConfig":
                    JsonObject configAsJson = new JsonObject();
                    configAsJson.addProperty("instanceURL", ConfigUtil.getSourcegraphUrl(this.project));
                    configAsJson.addProperty("isGlobbingEnabled", ConfigUtil.isGlobbingEnabled(this.project));
                    configAsJson.addProperty("accessToken", ConfigUtil.getAccessToken(this.project));
                    return createSuccessResponse(configAsJson);
                case "getTheme":
                    JsonObject currentThemeAsJson = ThemeUtil.getCurrentThemeAsJson();
                    return createSuccessResponse(currentThemeAsJson);
                case "saveLastSearch":
                    arguments = request.getAsJsonObject("arguments");
                    String query = arguments.get("query").getAsString();
                    boolean caseSensitive = arguments.get("caseSensitive").getAsBoolean();
                    String patternType = arguments.get("patternType").getAsString();
                    String selectedSearchContextSpec = arguments.get("selectedSearchContextSpec").getAsString();
                    ConfigUtil.setLastSearch(project, new Search(
                        query,
                        caseSensitive,
                        patternType,
                        selectedSearchContextSpec
                    ));
                    return createSuccessResponse(new JsonObject());
                case "loadLastSearch":
                    Search lastSearch = ConfigUtil.getLastSearch(this.project);

                    if (lastSearch == null) {
                        return createSuccessResponse(null);
                    }

                    JsonObject lastSearchAsJson = new JsonObject();
                    lastSearchAsJson.addProperty("query", lastSearch.getQuery());
                    lastSearchAsJson.addProperty("caseSensitive", lastSearch.isCaseSensitive());
                    lastSearchAsJson.addProperty("patternType", lastSearch.getPatternType());
                    lastSearchAsJson.addProperty("selectedSearchContextSpec", lastSearch.getSelectedSearchContextSpec());
                    return createSuccessResponse(lastSearchAsJson);
                case "preview":
                    arguments = request.getAsJsonObject("arguments");
                    previewContent = PreviewContent.fromJson(project, arguments);
                    previewPanel.setContent(previewContent);
                    return createSuccessResponse(null);
                case "clearPreview":
                    previewPanel.clearContent();
                    return createSuccessResponse(null);
                case "open":
                    arguments = request.getAsJsonObject("arguments");
                    previewContent = PreviewContent.fromJson(project, arguments);
                    try {
                        previewContent.openInEditorOrBrowser();
                    } catch (Exception e) {
                        return createErrorResponse("Error while opening link: " + e.getClass().getName() + ": " + e.getMessage(), convertStackTraceToString(e));
                    }
                    return createSuccessResponse(null);
                case "indicateFinishedLoading":
                    topPanel.setBrowserVisible(true);
                    return createSuccessResponse(null);
                default:
                    return createErrorResponse("Unknown action: '" + action + "'.", "No stack trace");
            }
        } catch (Exception e) {
            return createErrorResponse(action + ": " + e.getClass().getName() + ": " + e.getMessage(), convertStackTraceToString(e));
        }
    }

    public JSToJavaBridgeRequestHandler(@NotNull Project project, @NotNull PreviewPanel previewPanel, @NotNull BrowserAndLoadingPanel topPanel) {
        this.project = project;
        this.previewPanel = previewPanel;
        this.topPanel = topPanel;
    }

    public JBCefJSQuery.Response handleInvalidRequest(Exception e) {
        return createErrorResponse("Invalid JSON passed to bridge. The error is: " + e.getClass() + ": " + e.getMessage(), convertStackTraceToString(e));
    }

    @NotNull
    private JBCefJSQuery.Response createSuccessResponse(@Nullable JsonObject result) {
        return new JBCefJSQuery.Response(result != null ? result.toString() : "null");
    }

    @NotNull
    private JBCefJSQuery.Response createErrorResponse(@NotNull String errorMessage, @NotNull String stackTrace) {
        return new JBCefJSQuery.Response(null, 0, errorMessage + "\n" + stackTrace);
    }

    @NotNull
    private String convertStackTraceToString(@NotNull Exception e) {
        StringWriter sw = new StringWriter();
        PrintWriter pw = new PrintWriter(sw);
        e.printStackTrace(pw);
        return sw.toString();
    }
}
