import { ConfiguredExtension } from '@sourcegraph/extensions-client-common/lib/extensions/extension'
import {
    ConfigurationCascade,
    ConfigurationSubject,
    Settings,
} from '@sourcegraph/extensions-client-common/lib/settings'
import { Environment as CXPEnvironment } from 'cxp/module/environment/environment'

/** Client-side feature flag for using the new CXP controller and environment. */
export const USE_PLATFORM = localStorage.getItem('platform') !== null

/** React props or state representing the CXP environment. */
export interface CXPEnvironmentProps {
    /** The CXP environment. */
    cxpEnvironment: CXPEnvironment<ConfiguredExtension, ConfigurationCascade<ConfigurationSubject, Settings>>
}
