import { Controller as CXPController } from 'cxp/lib/environment/controller'
import { Environment as CXPEnvironment } from 'cxp/lib/environment/environment'
import { Extension as CXPExtension } from 'cxp/lib/environment/extension'
import { SourcegraphExtension } from '../schema/extension.schema'
import { ErrorLike } from '../util/errors'
import { CXPComponentProps } from './CXPComponent'
import { CXPRootProps } from './CXPRoot'

/** Client-side feature flag for using the new CXP controller and environment. */
export const USE_PLATFORM = localStorage.getItem('platform') !== null

/**
 * Adds the manifest to CXP extensions in the CXP environment, so we can consult it in the createMessageTransports
 * callback (to know how to communicate with or run the extension).
 */
export interface CXPExtensionWithManifest extends CXPExtension {
    isEnabled: boolean
    manifest: SourcegraphExtension | null | ErrorLike
}

/** React props or state representing the CXP environment. */
export interface CXPEnvironmentProps {
    /** The CXP environment. */
    cxpEnvironment: CXPEnvironment<CXPExtensionWithManifest>
}

/**
 * React props or state containing the CXP controller. There should be only a single CXP controller for the whole
 * application.
 */
export interface CXPControllerProps {
    /**
     * The CXP controller, which is used to communicate with the extensions and manages extensions based on the CXP
     * environment.
     */
    cxpController: CXPController<CXPExtensionWithManifest>
}

/** React props for components in the CXP environment. */
export interface CXPProps extends CXPEnvironmentProps, CXPControllerProps, CXPComponentProps, CXPRootProps {}
