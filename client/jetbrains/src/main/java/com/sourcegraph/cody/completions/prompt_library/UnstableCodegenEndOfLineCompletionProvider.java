package com.sourcegraph.cody.completions.prompt_library;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.sourcegraph.cody.api.Message;
import com.sourcegraph.cody.completions.UnstableCodegenLanguageUtil;
import com.sourcegraph.cody.vscode.CancellationToken;
import com.sourcegraph.cody.vscode.Completion;
import java.io.UnsupportedEncodingException;
import java.net.ConnectException;
import java.util.*;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.CompletableFuture;
import java.util.stream.Collectors;
import org.apache.http.HttpEntity;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

/** This is a rough implementation loosely translating unstable-codegen.ts */
public class UnstableCodegenEndOfLineCompletionProvider extends CompletionProvider {
  private final String fileName;
  private final String completionsEndpoint;
  private final String languageId;

  public UnstableCodegenEndOfLineCompletionProvider(
      SourcegraphNodeCompletionsClient completionsClient,
      int promptChars,
      int responseTokens,
      List<ReferenceSnippet> snippets,
      String prefix,
      String suffix,
      String injectPrefix,
      int defaultN,
      String fileName,
      @NotNull String completionsEndpoint,
      @Nullable String languageId) {
    super(
        completionsClient,
        promptChars,
        responseTokens,
        snippets,
        prefix,
        suffix,
        injectPrefix,
        defaultN);
    this.fileName = fileName;
    this.completionsEndpoint = completionsEndpoint;
    this.languageId = languageId;
  }

  @Override
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
          "<|"
              + UnstableCodegenLanguageUtil.getModelLanguageId(this.languageId, this.fileName)
              + "|>");
      params.put("prefix", this.prefix);
      params.put("suffix", this.suffix);
      params.put("top_p", 0.95);
      params.put("temperature", 0.2);
      params.put("max_tokens", 40);
      params.put("batch_size", makeEven(4));
      params.put("context", mapper.writeValueAsString(prepareContext(snippets, fileName)));
      params.put("completion_type", "automatic");

      StringEntity result = new StringEntity(mapper.writeValueAsString(params));
      result.setContentType("application/json");
      result.setContentEncoding("UTF-8");

      return result;
    } catch (JsonProcessingException | UnsupportedEncodingException e) {
      e.printStackTrace();
      return null;
    }
  }

  @Override
  public CompletableFuture<List<Completion>> generateCompletions(
      CancellationToken token, Optional<Integer> n) {
    return CompletableFuture.supplyAsync(
        () -> {
          StringEntity params = getParams();
          if (params == null) {
            System.err.println("Cody: Could not create params for unstable-codegen");
            return Collections.emptyList();
          }
          HttpPost httpPost = new HttpPost(completionsEndpoint);
          httpPost.setHeader("Content-Type", "application/json");
          httpPost.setHeader("Accept", "application/json");
          httpPost.setEntity(params);

          try (CloseableHttpClient client = HttpClients.createDefault()) {
            CloseableHttpResponse response = client.execute(httpPost);
            int responseCode = response.getStatusLine().getStatusCode();
            if (responseCode != 200) {
              System.err.println(
                  "Cody: `unstable-codegen` completion provider returned non-200 response code: "
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
            System.err.println(
                "Cody: Could not connect to the 'unstable-codegen' completion provider");
            return Collections.emptyList();
          } catch (Exception e) {
            e.printStackTrace();
            return Collections.emptyList();
          }
          return Collections.emptyList();
        });
  }

  private String postProcess(String content) {
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

  private Context prepareContext(List<ReferenceSnippet> snippets, String fileName) {
    List<Window> windows = new ArrayList<>();

    double similarity = 0.5;
    for (ReferenceSnippet snippet : snippets) {
      similarity *= 0.99;
      windows.add(new Window(snippet.filename, snippet.jaccard.text, similarity));
    }

    return new Context(fileName, windows);
  }

  class Context {
    private String currentFilePath;
    private List<Window> windows;

    public Context(String currentFilePath, List<Window> windows) {
      this.currentFilePath = currentFilePath;
      this.windows = windows;
    }

    public String getCurrentFilePath() {
      return currentFilePath;
    }

    public void setCurrentFilePath(String currentFilePath) {
      this.currentFilePath = currentFilePath;
    }

    public List<Window> getWindows() {
      return windows;
    }

    public void setWindows(List<Window> windows) {
      this.windows = windows;
    }

    // getters and setters
  }

  class Window {
    private String filePath;
    private String text;
    private double similarity;

    public Window(String filePath, String text, double similarity) {
      this.filePath = filePath;
      this.text = text;
      this.similarity = similarity;
    }

    public String getFilePath() {
      return filePath;
    }

    public void setFilePath(String filePath) {
      this.filePath = filePath;
    }

    public String getText() {
      return text;
    }

    public void setText(String text) {
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
