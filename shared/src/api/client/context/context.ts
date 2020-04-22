import { basename, dirname, extname } from 'path'
import { isSettingsValid, SettingsCascadeOrError } from '../../../settings/settings'
import { CodeEditorWithPartialModel } from '../services/editorService'

/**
 * Context is an arbitrary, immutable set of key-value pairs. Its value can be any JSON object.
 *
 * @template T If you have a value with a property of type T that is not one of the primitive types listed below
 * (or Context), you can use Context<T> to hold that value. T must be a value that can be represented by a JSON
 * object.
 */
export interface Context<T = never>
    extends Record<
        string,
        string | number | boolean | null | Context | T | (string | number | boolean | null | Context | T)[]
    > {}

export type ContributionScope =
    | CodeEditorWithPartialModel
    | {
          type: 'panelView'
          id: string
          hasLocations: boolean
      }

/**
 * Looks up a key in the computed context, which consists of computed context properties (with higher precedence)
 * and the context entries (with lower precedence).
 *
 * @param expr the context expr to evaluate
 * @param scope the user interface component in whose scope this computation should occur
 */
export function getComputedContextProperty(
    activeEditor: CodeEditorWithPartialModel | undefined,
    settings: SettingsCascadeOrError,
    context: Context<any>,
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
    const component: ContributionScope | null = scope || activeEditor || null
    if (key === 'resource' || key === 'component' /* BACKCOMPAT: allow 'component' */) {
        return !!component
    }
    if (key.startsWith('resource.')) {
        if (component?.type !== 'CodeEditor') {
            return null
        }
        // TODO(sqs): Define these precisely. If the resource is in a repository, what is the "path"? Is it the
        // path relative to the repository's root? If it's a file on disk, then "path" could also mean the
        // (absolute) path on the file system. Clear up that ambiguity.
        const prop = key.slice('resource.'.length)
        switch (prop) {
            case 'uri':
                return component.resource
            case 'basename':
                return basename(component.resource)
            case 'dirname':
                return dirname(component.resource)
            case 'extname':
                return extname(component.resource)
            case 'language':
                return component.model.languageId
            case 'type':
                return 'textDocument'
        }
    }
    if (key.startsWith('component.')) {
        if (component?.type !== 'CodeEditor') {
            return null
        }
        const prop = key.slice('component.'.length)
        switch (prop) {
            case 'type':
                return 'CodeEditor'
            case 'selections':
                return component.selections
            case 'selection':
                return component.selections[0] || null
            case 'selection.start':
                return component.selections[0] ? component.selections[0].start : null
            case 'selection.end':
                return component.selections[0] ? component.selections[0].end : null
            case 'selection.start.line':
                return component.selections[0] ? component.selections[0].start.line : null
            case 'selection.start.character':
                return component.selections[0] ? component.selections[0].start.character : null
            case 'selection.end.line':
                return component.selections[0] ? component.selections[0].end.line : null
            case 'selection.end.character':
                return component.selections[0] ? component.selections[0].end.character : null
        }
    }
    if (key.startsWith('panel.activeView.')) {
        if (component?.type !== 'panelView') {
            return null
        }
        const prop = key.slice('panel.activeView.'.length)
        switch (prop) {
            case 'id':
                return component.id
            case 'hasLocations':
                return component.hasLocations
        }
    }
    if (key === 'context') {
        return context
    }
    return context[key]
}
