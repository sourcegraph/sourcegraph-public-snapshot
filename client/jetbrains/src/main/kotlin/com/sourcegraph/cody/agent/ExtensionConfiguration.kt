package com.sourcegraph.cody.agent

data class ExtensionConfiguration(
    var serverEndpoint: String,
    var proxy: String? = null,
    var accessToken: String,
    var customHeaders: Map<String, String> = emptyMap(),
    var autocompleteAdvancedProvider: String? = null,
    var autocompleteAdvancedServerEndpoint: String? = null,
    var autocompleteAdvancedAccessToken: String? = null,
    var autocompleteAdvancedEmbeddings: Boolean = false,
    var debug: Boolean? = false,
    var verboseDebug: Boolean? = false,
    var codebase: String? = null
)
