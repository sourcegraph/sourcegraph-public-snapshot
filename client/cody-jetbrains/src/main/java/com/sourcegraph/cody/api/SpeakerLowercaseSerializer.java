package com.sourcegraph.cody.api;

import com.google.gson.JsonElement;
import com.google.gson.JsonPrimitive;
import com.google.gson.JsonSerializationContext;
import com.google.gson.JsonSerializer;
import java.lang.reflect.Type;

public class SpeakerLowercaseSerializer implements JsonSerializer<Speaker> {
  @Override
  public JsonElement serialize(Speaker src, Type typeOfSrc, JsonSerializationContext context) {
    return new JsonPrimitive(src.name().toLowerCase());
  }
}
