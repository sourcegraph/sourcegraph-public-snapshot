package com.sourcegraph.cody.api;

import com.google.gson.Gson;
import java.util.HashMap;
import java.util.Map;
import org.jetbrains.annotations.NotNull;

/** Wrapper for GraphQL requests that has a query and a list of variables. */
class GraphQLWrapper {
  @NotNull public String query;
  @NotNull public Map<String, Object> variables = new HashMap<>();

  public GraphQLWrapper(@NotNull String query) {
    this.query = query;
  }

  public GraphQLWrapper withVariable(@NotNull String key, @NotNull Object variable) {
    this.variables.put(key, variable);
    return this;
  }

  @NotNull
  public String toJsonString() {
    Gson gson = new Gson();
    return gson.toJson(this);
  }
}
