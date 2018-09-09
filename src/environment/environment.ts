import { Observable, of } from 'rxjs'
import { distinctUntilChanged, map } from 'rxjs/operators'
import { ConfigurationCascade } from '../protocol'
import { Range } from '../types/range'
import { Selection } from '../types/selection'
import { TextDocument, TextDocumentItem } from '../types/textDocument'
import { isEqual } from '../util'
import { Context, EMPTY_CONTEXT } from './context/context'
import { Extension } from './extension'

/**
 * A description of the environment represented by the Sourcegraph extension client application.
 *
 * This models the state of editor-like tools that display documents, allow selections and scrolling
 * in documents, and support extension configuration.
 *
 * @template X extension type, to support storing additional properties on extensions
 * @template C configuration cascade type
 */
export interface Environment<X extends Extension = Extension, C extends ConfigurationCascade = ConfigurationCascade> {
    /**
     * The active user interface component (i.e., file view or editor), or null if there is none.
     *
     * TODO(sqs): Support multiple components.
     */
    readonly component: Component | null

    /** The active extensions, or null if there are none. */
    readonly extensions: X[] | null

    /** The configuration cascade. */
    readonly configuration: C

    /** Arbitrary key-value pairs that describe other application state. */
    readonly context: Context
}

/** An empty Sourcegraph extension client environment. */
export const EMPTY_ENVIRONMENT: Environment<any, any> = {
    component: null,
    extensions: null,
    configuration: { merged: {} },
    context: EMPTY_CONTEXT,
}

/** An application component that displays a [TextDocument](#TextDocument). */
export interface Component {
    /** The document displayed by the component. */
    readonly document: TextDocumentItem

    /**
     * The selections in this component's document. If empty, there are no selections. The first element is
     * considered the primary selection (and some operations may only heed the primary selection).
     */
    readonly selections: Selection[]

    /** The vertical ranges in the document that are visible in the component. */
    readonly visibleRanges: Readonly<Range>[]
}

/**
 * Observables for changes to the environment.
 *
 * Includes derived observables for convenience.
 *
 * @template X extension type, to support storing additional properties on extensions
 * @template C configuration cascade type
 */
export interface ObservableEnvironment<X extends Extension, C extends ConfigurationCascade> {
    /** The environment (and changes to it). */
    readonly environment: Observable<Environment<X, C>>

    /** The environment's active component (and changes to it). */
    readonly component: Observable<Component | null>

    /** The active component's text document (and changes to it). */
    readonly textDocument: Observable<Pick<TextDocument, 'uri' | 'languageId'> | null>

    /** The environment's configuration cascade (and changes to it). */
    readonly configuration: Observable<C>

    /** The environment's context (and changes to it). */
    readonly context: Observable<Context>
}

/** An ObservableEnvironment that always represents the empty environment and never emits changes. */
export const EMPTY_OBSERVABLE_ENVIRONMENT: ObservableEnvironment<any, any> = {
    environment: of(EMPTY_ENVIRONMENT),
    component: of(null),
    textDocument: of(null),
    configuration: of({}),
    context: of(EMPTY_CONTEXT),
}

/**
 * Helper function for creating an ObservableEnvironment from the raw environment Observable.
 *
 * @template X extension type
 * @template C configuration cascade type
 */
export function createObservableEnvironment<X extends Extension, C extends ConfigurationCascade>(
    environment: Observable<Environment<X, C>>
): ObservableEnvironment<X, C> {
    const component = environment.pipe(
        map(({ component }) => component),
        distinctUntilChanged((a, b) => isEqual(a, b))
    )
    const textDocument = component.pipe(
        map(component => (component ? component.document : null)),
        distinctUntilChanged((a, b) => isEqual(a, b))
    )
    const configuration = environment.pipe(
        map(({ configuration }) => configuration),
        distinctUntilChanged((a, b) => isEqual(a, b))
    )
    const context = environment.pipe(
        map(({ context }) => context),
        distinctUntilChanged((a, b) => isEqual(a, b))
    )
    return {
        environment,
        component,
        textDocument,
        configuration,
        context,
    }
}
