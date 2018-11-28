import { basename, dirname, extname } from 'path'
import { isSettingsValid, SettingsCascadeOrError } from '../../../settings/settings'
import { Model } from '../model'
import { TextDocumentItem } from '../types/textDocument'

/**
 * Returns a new context created by applying the update context to the base context. It is equivalent to `{...base,
 * ...update}` in JavaScript except that null values in the update result in deletion of the property.
 */
export function applyContextUpdate(base: Context, update: Context): Context {
    const result = { ...base }
    for (const [key, value] of Object.entries(update)) {
        if (value === null) {
            delete result[key]
        } else {
            result[key] = value
        }
    }
    return result
}

/**
 * Context is an arbitrary, immutable set of key-value pairs.
 */
export interface Context {
    [key: string]: string | number | boolean | Context | null
}

export interface ContributionScope extends Pick<TextDocumentItem, 'uri' | 'languageId'> {}

/**
 * Looks up a key in the computed context, which consists of computed context properties (with higher precedence)
 * and the context entries (with lower precedence).
 *
 * @param expr the context expr to evaluate
 * @param scope the user interface component in whose scope this computation should occur
 */
export function getComputedContextProperty(
    model: Model,
    settings: SettingsCascadeOrError,
    context: Context,
    key: string,
    scope?: ContributionScope
): any {
    if (key.startsWith('config.')) {
        const prop = key.slice('config.'.length)
        const value = isSettingsValid(settings) ? settings.final[prop] : undefined
        // Map undefined to null because an undefined value is treated as "does not exist in
        // context" and an error is thrown, which is undesirable for config values (for
        // which a falsey null default is useful).
        return value === undefined ? null : value
    }
    const textDocument: ContributionScope | null =
        scope || (model.visibleTextDocuments && model.visibleTextDocuments[0])
    if (key === 'resource' || key === 'component' /* BACKCOMPAT: allow 'component' */) {
        return !!textDocument
    }
    if (key.startsWith('resource.')) {
        if (!textDocument) {
            return undefined
        }
        // TODO(sqs): Define these precisely. If the resource is in a repository, what is the "path"? Is it the
        // path relative to the repository's root? If it's a file on disk, then "path" could also mean the
        // (absolute) path on the file system. Clear up that ambiguity.
        const prop = key.slice('resource.'.length)
        switch (prop) {
            case 'uri':
                return textDocument.uri
            case 'basename':
                return basename(textDocument.uri)
            case 'dirname':
                return dirname(textDocument.uri)
            case 'extname':
                return extname(textDocument.uri)
            case 'language':
                return textDocument.languageId
            case 'type':
                return 'textDocument'
        }
    }
    if (key.startsWith('component.')) {
        if (!textDocument) {
            return undefined
        }
        const prop = key.slice('component.'.length)
        switch (prop) {
            case 'type':
                return 'textEditor'
        }
    }
    if (key === 'context') {
        return context
    }
    return context[key]
}
