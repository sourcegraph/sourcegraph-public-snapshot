package com.sourcegraph.cody.context

abstract class RepoAvailableEmbeddingStatus protected constructor(fullRepositoryName: String) :
    EmbeddingStatus {
  private val simpleRepositoryName: String

  init {
    simpleRepositoryName = getRepositoryNameAfterLastSlash(fullRepositoryName)
  }

  override fun getMainText(): String {
    return simpleRepositoryName
  }

  companion object {
    private fun getRepositoryNameAfterLastSlash(fullRepositoryName: String): String {
      var indexOfLastSlash = fullRepositoryName.lastIndexOf('/')
      val repoNameWithoutTrailingSlash =
          if (indexOfLastSlash == fullRepositoryName.length - 1)
              fullRepositoryName.substring(0, indexOfLastSlash)
          else fullRepositoryName
      indexOfLastSlash = repoNameWithoutTrailingSlash.lastIndexOf('/')
      return if (indexOfLastSlash != -1 &&
          indexOfLastSlash != repoNameWithoutTrailingSlash.length - 1)
          repoNameWithoutTrailingSlash.substring(indexOfLastSlash + 1)
      else repoNameWithoutTrailingSlash
    }
  }
}
