import { Environment } from '../../../../extensions-client-common/src/api/client/environment'
import { TextDocumentItem } from '../../../../extensions-client-common/src/api/client/types/textDocument'
import { ConfiguredExtension } from '../../../../extensions-client-common/src/extensions/extension'
import { Settings, SettingsCascade, SettingsSubject } from '../../../../extensions-client-common/src/settings'

/** React props or state representing the Sourcegraph extensions environment. */
export interface ExtensionsEnvironmentProps {
    /** The Sourcegraph extensions environment. */
    extensionsEnvironment: Environment<ConfiguredExtension, SettingsCascade<SettingsSubject, Settings>>
}

/** React props for components that participate in the Sourcegraph extensions environment. */
export interface ExtensionsDocumentsProps {
    /**
     * Called when the Sourcegraph extensions environment's set of visible text documents changes.
     */
    extensionsOnVisibleTextDocumentsChange: (visibleTextDocuments: TextDocumentItem[] | null) => void
}
