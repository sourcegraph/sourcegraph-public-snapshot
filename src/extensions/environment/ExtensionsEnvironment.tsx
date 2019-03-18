import { ConfiguredExtension } from '@sourcegraph/extensions-client-common/lib/extensions/extension'
import {
    ConfigurationCascade,
    ConfigurationSubject,
    Settings,
} from '@sourcegraph/extensions-client-common/lib/settings'
import { Environment } from 'sourcegraph/module/client/environment'
import { TextDocumentItem } from 'sourcegraph/module/client/types/textDocument'

/** React props or state representing the Sourcegraph extensions environment. */
export interface ExtensionsEnvironmentProps {
    /** The Sourcegraph extensions environment. */
    extensionsEnvironment: Environment<ConfiguredExtension, ConfigurationCascade<ConfigurationSubject, Settings>>
}

/** React props for components that participate in the Sourcegraph extensions environment. */
export interface ExtensionsDocumentsProps {
    /**
     * Called when the Sourcegraph extensions environment's set of visible text documents changes.
     */
    extensionsOnVisibleTextDocumentsChange: (visibleTextDocuments: TextDocumentItem[] | null) => void
}
