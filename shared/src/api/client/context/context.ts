import { Position, Selection } from '@sourcegraph/extension-api-types'
import { basename, dirname, extname } from 'path'
import { isSettingsValid, SettingsCascadeOrError } from '../../../settings/settings'
import { ViewerWithPartialModel } from '../services/viewerService'

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
        string | number | boolean | null | Context<T> | T | (string | number | boolean | null | Context<T> | T)[]
    > {}

type ValueOf<T> = T[keyof T]

export type ContributionScope =
    | ViewerWithPartialModel
    | {
          type: 'panelView'
          id: string
          hasLocations: boolean
      }

/** The types of the builtin context keys (such as `component.selections`). */
type BuiltinContextValuesTypes = Selection | Selection[] | Position

/**
 * Looks up a key in the computed context, which consists of computed context properties (with higher precedence)
 * and the context entries (with lower precedence).
 *
 * @param expr the context expr to evaluate
 * @param scope the user interface component in whose scope this computation should occur
 */
export function getComputedContextProperty<T>(
    activeEditor: ViewerWithPartialModel | undefined,
    settings: SettingsCascadeOrError,
    context: Context<T>,
    key: string,
    scope?: ContributionScope
): ValueOf<Context<BuiltinContextValuesTypes | T>> {
    const data: Context<BuiltinContextValuesTypes | T> = { ...context }

    // Settings (`config.` prefix)
    if (isSettingsValid(settings)) {
        for (const [key, value] of Object.entries<ValueOf<Context>>(settings.final)) {
            data[`config.${key}`] = value
        }
    }

    // Resource (`resource.` prefix)
    const component: ContributionScope | null = scope || activeEditor || null
    data.resource = Boolean(component)
    data.component = Boolean(component) // BACKCOMPAT: allow 'component' key
    if (component?.type === 'CodeEditor') {
        // TODO(sqs): Define these precisely. If the resource is in a repository, what is the "path"? Is it the
        // path relative to the repository's root? If it's a file on disk, then "path" could also mean the
        // (absolute) path on the file system. Clear up that ambiguity.
        data['resource.uri'] = component.resource
        data['resource.basename'] = basename(component.resource)
        data['resource.dirname'] = dirname(component.resource)
        data['resource.extname'] = extname(component.resource)
        data['resource.language'] = component.model.languageId
        data['resource.type'] = 'textDocument'

        data['component.type'] = 'CodeEditor'
        data['component.selections'] = component.selections
        data['component.selection'] = component.selections[0] || null
        data['component.selection.start'] = component.selections[0] ? component.selections[0].start : null
        data['component.selection.end'] = component.selections[0] ? component.selections[0].end : null
        data['component.selection.start.line'] = component.selections[0] ? component.selections[0].start.line : null
        data['component.selection.start.character'] = component.selections[0]
            ? component.selections[0].start.character
            : null
        data['component.selection.end.line'] = component.selections[0] ? component.selections[0].end.line : null
        data['component.selection.end.character'] = component.selections[0]
            ? component.selections[0].end.character
            : null
    }

    // Panel (`panel.` prefix)
    if (component?.type === 'panelView') {
        data['panel.activeView.id'] = component.id
        data['panel.activeView.hasLocations'] = component.hasLocations
    }

    data.context = context

    // BACKCOMPAT: If the key is not found, we return null, not undefined.
    const value = data[key]
    return value === undefined ? null : value
}
