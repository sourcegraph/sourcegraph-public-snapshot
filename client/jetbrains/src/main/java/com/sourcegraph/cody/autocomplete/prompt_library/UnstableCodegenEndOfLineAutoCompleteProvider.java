package com.sourcegraph.cody.autocomplete.prompt_library;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.intellij.openapi.diagnostic.Logger;
import com.sourcegraph.cody.api.Message;
import com.sourcegraph.cody.autocomplete.UnstableCodegenLanguageUtil;
import com.sourcegraph.cody.vscode.CancellationToken;
import com.sourcegraph.cody.vscode.Completion;
import com.sourcegraph.cody.vscode.TextDocument;
import java.io.UnsupportedEncodingException;
import java.net.ConnectException;
import java.util.*;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.CompletableFuture;
import java.util.stream.Collectors;
import org.apache.http.HttpEntity;
import org.apache.http.client.config.CookieSpecs;
import org.apache.http.client.config.RequestConfig;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

/** This is a rough implementation loosely translating unstable-codegen.ts */
public class UnstableCodegenEndOfLineAutoCompleteProvider extends AutoCompleteProvider {
  private static final Logger logger =
      Logger.getInstance(UnstableCodegenEndOfLineAutoCompleteProvider.class);
  @NotNull private final String autocompleteEndpoint;
  @NotNull private final TextDocument textDocument;

  public UnstableCodegenEndOfLineAutoCompleteProvider(
      @NotNull List<ReferenceSnippet> snippets,
      @NotNull String prefix,
      @NotNull String suffix,
      @NotNull String autocompleteEndpoint,
      @NotNull TextDocument textDocument) {
    super(
        null, // unused
        -1, // unused
        -1,
        snippets,
        prefix,
        suffix,
        "", // unused
        -1 // unused
        );
    this.autocompleteEndpoint = autocompleteEndpoint;
    this.textDocument = textDocument;
  }

  @Override
  @NotNull
  protected List<Message> createPromptPrefix() {
    // it seems it's not necessary for unstable-codegen for now
    return Collections.emptyList();
  }

  @Nullable
  private StringEntity getParams() {
    try {
      ObjectMapper mapper = new ObjectMapper();
      Map<String, Object> params = new HashMap<>();
      params.put("debug_ext_path", "cody");
      params.put(
          "lang_prefix",
          "<|" + UnstableCodegenLanguageUtil.getModelLanguageId(textDocument) + "|>");
      params.put("prefix", this.prefix);
      params.put("suffix", this.suffix);
      params.put("top_p", 0.95);
      params.put("temperature", 0.2);
      params.put("max_tokens", 40);
      params.put("batch_size", makeEven(4));
      params.put(
          "context", mapper.writeValueAsString(prepareContext(snippets, textDocument.fileName())));
      params.put("completion_type", "automatic");

      StringEntity result = new StringEntity(mapper.writeValueAsString(params));
      result.setContentType("application/json");
      result.setContentEncoding("UTF-8");

      return result;
    } catch (JsonProcessingException | UnsupportedEncodingException e) {
      logger.error(e);
      return null;
    }
  }

  @Override
  @NotNull
  public CompletableFuture<List<Completion>> generateCompletions(
      @NotNull CancellationToken token, @NotNull Optional<Integer> n) {
    return CompletableFuture.supplyAsync(
        () -> {
          StringEntity params = getParams();
          if (params == null) {
            logger.error("Cody: Could not create params for unstable-codegen");
            return Collections.emptyList();
          }
          HttpPost httpPost = new HttpPost(autocompleteEndpoint);
          httpPost.setHeader("Content-Type", "application/json");
          httpPost.setHeader("Accept", "application/json");
          httpPost.setEntity(params);

          try (CloseableHttpClient client =
              HttpClients.custom()
                  .setDefaultRequestConfig(
                      RequestConfig.custom().setCookieSpec(CookieSpecs.STANDARD).build())
                  .build()) {
            CloseableHttpResponse response = client.execute(httpPost);
            int responseCode = response.getStatusLine().getStatusCode();
            if (responseCode != 200) {
              logger.error(
                  "Cody: `unstable-codegen` autocomplete provider returned non-200 response code: "
                      + responseCode);
              return Collections.emptyList();
            }
            HttpEntity responseEntity = response.getEntity();

            if (responseEntity != null) {
              String responseString = EntityUtils.toString(responseEntity);
              ObjectMapper mapper = new ObjectMapper();
              JsonNode rootNode = mapper.readTree(responseString);
              JsonNode completionsNode = rootNode.get("completions");

              List<String> completions = new ArrayList<>();
              for (JsonNode completionNode : completionsNode) {
                String completion = completionNode.get("completion").asText();
                completions.add(completion);
              }

              return completions.stream()
                  .map(this::postProcess)
                  .map(
                      c -> new Completion(prefix, Collections.emptyList(), this.postProcess(c), ""))
                  .collect(Collectors.toList());
            }
          } catch (ConnectException e) {
            logger.error("Cody: Could not connect to the 'unstable-codegen' autocomplete provider");
            return Collections.emptyList();
          } catch (Exception e) {
            logger.error(e);
            return Collections.emptyList();
          }
          return Collections.emptyList();
        },
        executor);
  }

  @NotNull
  private String postProcess(@NotNull String content) {
    if (content.contains("\n")) {
      return content.substring(0, content.indexOf('\n')).trim();
    } else return content.trim();
  }

  private int makeEven(int number) {
    if (number % 2 == 1) {
      return number + 1;
    }
    return number;
  }

  @NotNull
  private Context prepareContext(
      @NotNull List<ReferenceSnippet> snippets, @NotNull String fileName) {
    List<Window> windows = new ArrayList<>();

    double similarity = 0.5;
    for (ReferenceSnippet snippet : snippets) {
      similarity *= 0.99;
      windows.add(new Window(snippet.filename, snippet.jaccard.text, similarity));
    }

    return new Context(fileName, windows);
  }

  static class Context {
    @JsonProperty("current_file_path")
    @NotNull
    private String currentFilePath;

    @NotNull private List<Window> windows;

    public Context(@NotNull String currentFilePath, @NotNull List<Window> windows) {
      this.currentFilePath = currentFilePath;
      this.windows = windows;
    }

    public @NotNull String getCurrentFilePath() {
      return currentFilePath;
    }

    public void setCurrentFilePath(@NotNull String currentFilePath) {
      this.currentFilePath = currentFilePath;
    }

    public @NotNull List<Window> getWindows() {
      return windows;
    }

    public void setWindows(@NotNull List<Window> windows) {
      this.windows = windows;
    }

    // getters and setters
  }

  static class Window {
    @JsonProperty("file_path")
    @NotNull
    private String filePath;

    @NotNull private String text;
    private double similarity;

    public Window(@NotNull String filePath, @NotNull String text, double similarity) {
      this.filePath = filePath;
      this.text = text;
      this.similarity = similarity;
    }

    public @NotNull String getFilePath() {
      return filePath;
    }

    public void setFilePath(@NotNull String filePath) {
      this.filePath = filePath;
    }

    public @NotNull String getText() {
      return text;
    }

    public void setText(@NotNull String text) {
      this.text = text;
    }

    public double getSimilarity() {
      return similarity;
    }

    public void setSimilarity(double similarity) {
      this.similarity = similarity;
    }
  }
}
