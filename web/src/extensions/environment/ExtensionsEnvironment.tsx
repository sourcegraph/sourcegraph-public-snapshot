import { ConfiguredExtension } from '@sourcegraph/extensions-client-common/lib/extensions/extension'
import {
    ConfigurationCascade,
    ConfigurationSubject,
    Settings,
} from '@sourcegraph/extensions-client-common/lib/settings'
import { Component, Environment } from '@sourcegraph/sourcegraph.proposed/module/environment/environment'

/**
 * Whether the platform (Sourcegraph extensions and the extension registry) should be enabled for the viewer.
 *
 * On local dev instances, run `localStorage.platform=true;location.reload()` to enable this.
 *
 * The server (site config experimentalFeatures.platform value) and browser (localStorage.platform) feature flags
 * must be enabled for this to be true.
 *
 */
export const USE_PLATFORM = !!window.context.platformEnabled && localStorage.getItem('platform') !== null

/** React props or state representing the Sourcegraph extensions environment. */
export interface ExtensionsEnvironmentProps {
    /** The Sourcegraph extensions environment. */
    extensionsEnvironment: Environment<ConfiguredExtension, ConfigurationCascade<ConfigurationSubject, Settings>>
}

/** React props for components that participate in the Sourcegraph extensions environment. */
export interface ExtensionsComponentProps {
    /**
     * Called when the Sourcegraph extensions environment component changes.
     */
    extensionsOnComponentChange: (component: Component | null) => void
}
