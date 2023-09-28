package com.sourcegraph.cody.agent.protocol

import com.sourcegraph.cody.agent.ExtensionConfiguration

data class ClientInfo(
    var name: String,
    var version: String,
    var workspaceRootPath: String? = null,
    var extensionConfiguration: ExtensionConfiguration? = null
)
