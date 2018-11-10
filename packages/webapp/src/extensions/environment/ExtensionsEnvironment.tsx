import { Environment } from '@sourcegraph/extensions-client-common/lib/api/client/environment'
import { TextDocumentItem } from '@sourcegraph/extensions-client-common/lib/api/client/types/textDocument'
import { ConfiguredExtension } from '@sourcegraph/extensions-client-common/lib/extensions/extension'
import { Settings, SettingsCascade, SettingsSubject } from '@sourcegraph/extensions-client-common/lib/settings'

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
