package com.sourcegraph.cody.api

import com.fasterxml.jackson.annotation.JsonAutoDetect
import com.fasterxml.jackson.annotation.JsonInclude
import com.fasterxml.jackson.core.JsonParseException
import com.fasterxml.jackson.core.JsonProcessingException
import com.fasterxml.jackson.databind.DeserializationFeature
import com.fasterxml.jackson.databind.JavaType
import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.databind.PropertyNamingStrategies
import com.fasterxml.jackson.databind.SerializationFeature
import com.fasterxml.jackson.databind.introspect.VisibilityChecker
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import java.awt.Image
import java.io.IOException
import java.io.InputStream
import java.io.Reader
import java.text.SimpleDateFormat
import java.util.TimeZone
import javax.imageio.ImageIO

object SourcegraphApiContentHelper {
  const val JSON_MIME_TYPE = "application/json"

  private val jackson: ObjectMapper =
      jacksonObjectMapper()
          .genericConfig()
          .setPropertyNamingStrategy(PropertyNamingStrategies.SNAKE_CASE)

  private val gqlJackson: ObjectMapper =
      jacksonObjectMapper()
          .genericConfig()
          .setPropertyNamingStrategy(PropertyNamingStrategies.LOWER_CAMEL_CASE)

  private fun ObjectMapper.genericConfig(): ObjectMapper =
      this.setDateFormat(SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ssXXX"))
          .setTimeZone(TimeZone.getDefault())
          .configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false)
          .configure(SerializationFeature.FAIL_ON_EMPTY_BEANS, false)
          .setSerializationInclusion(JsonInclude.Include.NON_NULL)
          .setVisibility(
              VisibilityChecker.Std(
                  JsonAutoDetect.Visibility.NONE,
                  JsonAutoDetect.Visibility.NONE,
                  JsonAutoDetect.Visibility.NONE,
                  JsonAutoDetect.Visibility.NONE,
                  JsonAutoDetect.Visibility.ANY))

  @JvmStatic
  @Throws(SourcegraphJsonException::class)
  fun <T> fromJson(string: String, clazz: Class<T>, gqlNaming: Boolean = false): T {
    try {
      return getObjectMapper(gqlNaming).readValue(string, clazz)
    } catch (e: JsonParseException) {
      throw SourcegraphJsonException("Can't parse Sourcegraph response", e)
    }
  }

  @JvmStatic
  @Throws(SourcegraphJsonException::class)
  fun <T> readJsonObject(
      reader: Reader,
      clazz: Class<T>,
      vararg parameters: Class<*>,
      gqlNaming: Boolean = false
  ): T {
    return readJson(
        reader, jackson.typeFactory.constructParametricType(clazz, *parameters), gqlNaming)
  }

  @Throws(SourcegraphJsonException::class)
  private fun <T> readJson(reader: Reader, type: JavaType, gqlNaming: Boolean = false): T {
    try {
      @Suppress("UNCHECKED_CAST")
      if (type.isTypeOrSubTypeOf(Unit::class.java) || type.isTypeOrSubTypeOf(Void::class.java))
          return Unit as T
      return getObjectMapper(gqlNaming).readValue(reader, type)
    } catch (e: JsonProcessingException) {
      throw SourcegraphJsonException("Can't parse Sourcegraph response", e)
    }
  }

  @JvmStatic
  @Throws(SourcegraphJsonException::class)
  fun toJson(content: Any, gqlNaming: Boolean = false): String {
    try {
      return getObjectMapper(gqlNaming).writeValueAsString(content)
    } catch (e: JsonProcessingException) {
      throw SourcegraphJsonException("Can't serialize Sourcegraph request body", e)
    }
  }

  private fun getObjectMapper(gqlNaming: Boolean = false): ObjectMapper =
      if (!gqlNaming) jackson else gqlJackson

  @JvmStatic
  @Throws(IOException::class)
  fun loadImage(stream: InputStream): Image {
    return ImageIO.read(stream)
  }
}
