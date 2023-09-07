package com.sourcegraph.cody.api

import com.intellij.util.ThrowableConvertor
import java.io.IOException
import java.io.InputStream
import java.io.Reader

interface SourcegraphApiResponse {
    fun findHeader(headerName: String): String?

    @Throws(IOException::class)
    fun <T> readBody(converter: ThrowableConvertor<Reader, T, IOException>): T

    @Throws(IOException::class)
    fun <T> handleBody(converter: ThrowableConvertor<InputStream, T, IOException>): T
}
