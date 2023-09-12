package com.sourcegraph.cody.auth

/**
 * Base class for an account which correspond to a certain server Most systems can have multiple
 * servers while some have only the central one
 *
 * @property server some definition of a server, which can be presented to a user
 */
abstract class ServerAccount : Account() {
  abstract val server: ServerPath
}
