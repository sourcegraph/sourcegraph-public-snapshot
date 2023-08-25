package com.sourcegraph.cody.config

import com.google.common.annotations.VisibleForTesting
import com.intellij.collaboration.api.graphql.CachingGraphQLQueryLoader
import com.intellij.util.io.isDirectory
import java.nio.file.Files
import java.nio.file.Paths
import java.util.stream.Collectors

object SourcegraphGQLQueryLoader: CachingGraphQLQueryLoader() {
    @VisibleForTesting
    fun findAllQueries(): List<String> {
        val url = SourcegraphGQLQueryLoader::class.java.classLoader.getResource("graphql/query")!!
        val directory = Paths.get(url.toURI())
        return Files.walk(directory)
            .filter { !it.isDirectory() }
            .map { "graphql/query/" + it.fileName.toString() }
            .collect(Collectors.toList())
    }
}
