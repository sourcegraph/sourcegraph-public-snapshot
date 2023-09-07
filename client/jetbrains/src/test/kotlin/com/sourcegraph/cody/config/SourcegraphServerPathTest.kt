package com.sourcegraph.cody.config

import com.sourcegraph.config.ConfigUtil
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class SourcegraphServerPathTest {

  @Test
  fun `should create server path for dotcom`() {
    // given
    val url = ConfigUtil.DOTCOM_URL

    // when
    val serverPath = SourcegraphServerPath.from(url, "")

    // then
    assertThat(serverPath.url).isEqualTo(url)
  }

  @Test
  fun `should create server path with additional slash at the end if missing`() {
    // given
    val url = "https://sourcegraph.com"

    // then
    val serverPath = SourcegraphServerPath.from(url, "")

    // then
    assertThat(serverPath.url).isEqualTo("https://sourcegraph.com/")
  }

  @Test
  fun `should create server path with additional https schema if it's missing`() {
    // given
    val url = "sourcegraph.com"

    // then
    val serverPath = SourcegraphServerPath.from(url, "")

    // then
    assertThat(serverPath.url).isEqualTo("https://sourcegraph.com/")
  }

  @Test
  fun `should create server path with port`() {
    // given
    val url = "sourcegraph.com:80"

    // then
    val serverPath = SourcegraphServerPath.from(url, "")

    // then
    assertThat(serverPath.url).isEqualTo("https://sourcegraph.com:80/")
  }

  @Test
  fun `should create server path with url path`() {
    // given
    val url = "sourcegraph.com:80/some/path"

    // then
    val serverPath = SourcegraphServerPath.from(url, "")

    // then
    assertThat(serverPath.url).isEqualTo("https://sourcegraph.com:80/some/path/")
  }
}
