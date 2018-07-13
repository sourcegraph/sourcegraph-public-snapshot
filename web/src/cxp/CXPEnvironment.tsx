import { Environment as CXPEnvironment } from 'cxp/lib/environment/environment'
import { Extension as CXPExtension } from 'cxp/lib/environment/extension'
import { ConfiguredExtension } from '../extensions/extension'
import { CXPComponentProps } from './CXPComponent'
import { CXPRootProps } from './CXPRoot'

/** Feature flag for using the new CXP controller and environment. */
export const USE_CXP = localStorage.getItem('cxp') !== null

/**
 * Adds the manifest to CXP extensions in the CXP environment, so we can consult it in the createMessageTransports
 * callback (to know how to communicate with or run the extension).
 */
export interface CXPExtensionWithManifest extends CXPExtension {
    manifest: ConfiguredExtension['manifest']
}

/** React props or state representing the CXP environment. */
export interface CXPEnvironmentProps {
    /** The CXP environment. */
    cxpEnvironment: CXPEnvironment<CXPExtensionWithManifest>
}

/** React props for components in the CXP environment. */
export interface CXPProps extends CXPEnvironmentProps, CXPComponentProps, CXPRootProps {}
