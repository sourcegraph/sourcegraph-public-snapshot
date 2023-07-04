package com.sourcegraph.cody.context;

import static org.assertj.core.api.Assertions.assertThat;
import static org.junit.jupiter.api.Assertions.*;

import java.util.stream.Stream;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

class RepositoryIndexedEmbeddingStatusTest {

  @ParameterizedTest
  @MethodSource("repositoryNames")
  public void shouldReturnSimpleRepositoryNameAsMainText(
      String fullRepositoryName, String expectedRepositoryName) {
    // given
    RepositoryIndexedEmbeddingStatus status =
        new RepositoryIndexedEmbeddingStatus(fullRepositoryName);

    // when
    String mainText = status.getMainText();

    // then
    assertThat(mainText).isEqualTo(expectedRepositoryName);
  }

  public static Stream<Arguments> repositoryNames() {
    return Stream.of(
        Arguments.of("sourcegraph", "sourcegraph"),
        Arguments.of("sourcegraph", "sourcegraph"),
        Arguments.of("github.com/sourcegraph/", "sourcegraph"),
        Arguments.of("github.com/sourcegraph/cody", "cody"));
  }
}
