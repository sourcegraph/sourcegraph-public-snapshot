import { ConfiguredExtension } from '@sourcegraph/extensions-client-common/lib/extensions/extension'
import {
    ConfigurationCascade,
    ConfigurationSubject,
    Settings,
} from '@sourcegraph/extensions-client-common/lib/settings'
import { Component, Environment as CXPEnvironment } from 'cxp/module/environment/environment'

/**
 * Whether the platform (CXP extensions and the extension registry) should be enabled for the viewer.
 *
 * On local dev instances, run `localStorage.platform=true;location.reload()` to enable this.
 *
 * The server (site config experimentalFeatures.platform value) and browser (localStorage.platform) feature flags
 * must be enabled for this to be true.
 *
 */
export const USE_PLATFORM = !!window.context.platformEnabled && localStorage.getItem('platform') !== null

/** React props or state representing the CXP environment. */
export interface CXPEnvironmentProps {
    /** The CXP environment. */
    cxpEnvironment: CXPEnvironment<ConfiguredExtension, ConfigurationCascade<ConfigurationSubject, Settings>>
}

/** React props for components that participate in the CXP environment. */
export interface CXPComponentProps {
    /**
     * Called when the CXP component changes.
     */
    cxpOnComponentChange: (component: Component | null) => void
}
