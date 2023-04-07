package com.sourcegraph.jody;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import com.sourcegraph.api.GraphQlClient;
import org.kohsuke.rngom.util.Uri;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class CompletionsService {

    private final String instanceUrl;
    private final String token;

    public CompletionsService(String instanceUrl, String token) {
        this.instanceUrl = instanceUrl;
        this.token = token;
    }

    public String send(CompletionsInput input) {


        try {
            Gson gson = new Gson();

            String query = "query comp($input:CompletionsInput!) {completions(input:$input)}";
            var wrapper = new GraphQLWrapper(query).withVariable("input", input);
            if (wrapper == null) {
                return null;
            }
            var json = gson.toJson(wrapper);
            if (json == null) {
                return null;
            }
            System.out.println(json);

            URI uri = URI.create(instanceUrl);
            if (uri == null) {
                return null;
            }
            var body = HttpClient.newHttpClient().send(HttpRequest.newBuilder(uri)
                .POST(HttpRequest.BodyPublishers.ofString(json))
                .header("Authorization", "token " + token)
                .build(), HttpResponse.BodyHandlers.ofString())
                .body();
            if (body == null) {
                return null;
            }

            return gson.fromJson(body, JsonObject.class).getAsJsonObject("data").getAsJsonPrimitive("completions").getAsString();
        } catch (Exception ex) {
            ex.printStackTrace();
        }
        return "";
    }
}

class GraphQLWrapper {
    public String query;
    public Map<String, Object> variables = new HashMap<>();

    public GraphQLWrapper(String query) {
        this.query = query;
    }

    public GraphQLWrapper withVariable(String key, Object variable) {
        this.variables.put(key, variable);
        return this;
    }
}

class CompletionsInput {
    private List<Message> messages;
    private float temperature;
    private int maxTokensToSample;
    private int topK;
    private int topP;

    public List<Message> getMessages() {
        return messages;
    }

    public float getTemperature() {
        return temperature;
    }

    public int getMaxTokensToSample() {
        return maxTokensToSample;
    }

    public int getTopK() {
        return topK;
    }

    public int getTopP() {
        return topP;
    }

    static CompletionsInputBuilder builder() {
        return new CompletionsInputBuilder();
    }

    static class CompletionsInputBuilder {
        private final List<Message> messages = new ArrayList<>();
        private float temperature;
        private int maxTokensToSample;
        private int topK;
        private int topP;

        public CompletionsInputBuilder addMessage(SpeakerType speaker, String text) {
            messages.add(new Message(speaker, text));
            return this;
        }

        public CompletionsInputBuilder setTemperature(float temperature) {
            this.temperature = temperature;
            return this;
        }

        public CompletionsInputBuilder toJson() {
            Gson gson = new Gson();
            return this;
        }

        public CompletionsInputBuilder setMaxTokensToSample(int maxTokensToSample) {
            this.maxTokensToSample = maxTokensToSample;
            return this;
        }

        public CompletionsInputBuilder setTopK(int topK) {
            this.topK = topK;
            return this;
        }

        public CompletionsInputBuilder setTopP(int topP) {
            this.topP = topP;
            return this;
        }

        public CompletionsInput build() {
            CompletionsInput input = new CompletionsInput();
            input.messages = messages;
            input.temperature = temperature;
            input.maxTokensToSample = maxTokensToSample;
            input.topK = topK;
            input.topP = topP;
            return input;
        }
    }
}

class Message {
    private SpeakerType speaker;
    private String text;

    public Message(SpeakerType speaker, String text) {
        this.speaker = speaker;
        this.text = text;
    }

    public SpeakerType getSpeaker() {
        return speaker;
    }

    public String getText() {
        return text;
    }

    @Override
    public String toString() {
        return "Message{" +
            "speaker=" + speaker +
            ", text='" + text + '\'' +
            '}';
    }
}

enum SpeakerType {
    HUMAN,
    ASSISTANT
}


