import type { DocumentHighlightKind as LegacyDocumentHighlightKind } from '../../../codeintel/legacy-extensions/api'

/**
 * The type of a document highlight.
 * This is needed because if DocumentHighlightKind enum values are referenced,
 * the `sourcegraph` module import at the top of the file is emitted in the generated code.
 */
export const DocumentHighlightKind: typeof LegacyDocumentHighlightKind = {
    Text: 'text' as LegacyDocumentHighlightKind.Text,
    Read: 'read' as LegacyDocumentHighlightKind.Read,
    Write: 'write' as LegacyDocumentHighlightKind.Write,
}
