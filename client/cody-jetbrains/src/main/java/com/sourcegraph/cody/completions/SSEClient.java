package com.sourcegraph.cody.completions;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import com.google.gson.JsonPrimitive;
import org.apache.commons.io.IOUtils;
import org.apache.http.HttpStatus;
import org.jetbrains.annotations.NotNull;

import java.io.*;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.net.http.HttpResponse.BodyHandlers;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.util.Objects;

public class SSEClient {
    private final String url;
    private final String accessToken;
    private final String body;
    private final CompletionsCallbacks cb;

    private InputStream inputStream;

    public SSEClient(@NotNull String url, @NotNull String accessToken, @NotNull String body, @NotNull CompletionsCallbacks cb) {
        this.url = url;
        this.body = body;
        this.accessToken = accessToken;
        this.cb = cb;
    }

    public void start() {
        try {
            HttpRequest.Builder requestBuilder = HttpRequest.newBuilder()
                .uri(new URI(url))
                .version(HttpClient.Version.HTTP_2)
                .header("Content-Type", "application/json; charset=utf-8")
                .header("X-Sourcegraph-Should-Trace", "false")
                .header("Accept", "text/event-stream")
                .header("Cache-Control", "no-cache")
                .header("Authorization", "token " + accessToken)
                .timeout(Duration.ofSeconds(30))
                .POST(HttpRequest.BodyPublishers.ofString(body));
            HttpRequest request = requestBuilder.build();

            HttpResponse<InputStream> response = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(30))
                .build()
                .send(request, BodyHandlers.ofInputStream());

            if (response.statusCode() == HttpStatus.SC_OK) {
                cb.onSubscribed();
                inputStream = response.body();
                handleResponse(inputStream);
            } else {
                String result;
                try (InputStream ignored = response.body()) {
                    result = IOUtils.toString(response.body(), StandardCharsets.UTF_8);
                }
                cb.onError(new Error("Got error response " + response.statusCode() + ": " + result));
            }
        } catch (Exception e) {
            if (e.getCause() instanceof InterruptedException) {
                cb.onError(e); // TODO: Handle interruptions
            } else {
                cb.onError(e);
            }
        }
    }

    public void stopCurrentRequest() {
        if (inputStream != null) {
            try {
                inputStream.close();
            } catch (Exception e) {
                System.err.println("Got error stopCurrentRequest: " + e.getMessage());
            }
        }
    }

    private void handleResponse(@NotNull InputStream inputStream) {
        /*
         * Streams go like this:
         * event: completion
         * data: Hello
         *
         * event: completion
         * data: Hello, there
         *
         * event: done
         * data:
         */
        String eventName = null;
        try (BufferedInputStream in = IOUtils.buffer(inputStream)) {
            try (BufferedReader reader = new BufferedReader(new InputStreamReader(in, StandardCharsets.UTF_8))) {
                String line;
                StringBuilder messageBuilder = new StringBuilder();
                while ((line = reader.readLine()) != null) {
                    if (line.startsWith("data:")) {
                        messageBuilder.append(line.substring(5));
                    } else if (line.startsWith("event:")) {
                        eventName = line.substring(6).trim();
                    }
                    if (line.trim().isEmpty() && messageBuilder.length() > 0) {
                        String message = messageBuilder.toString();
                        if (Objects.equals(eventName, "completion")) { // Completion
                            JsonObject json = new Gson().fromJson(message, JsonObject.class);
                            JsonPrimitive completion = json.getAsJsonPrimitive("completion");
                            cb.onData(completion != null ? completion.getAsString().trim() : null);
                        } else if (Objects.equals(eventName, "done")) { // Done
                            stopCurrentRequest();
                            cb.onComplete();
                            return;
                        }
                        messageBuilder = new StringBuilder();
                    }
                }
                if (messageBuilder.length() > 0) {
                    System.out.println("Non-processed data: {}" + messageBuilder);
                }
            }
        } catch (Exception e) {
            cb.onError(e);
        }
    }
}
