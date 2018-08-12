import { Observable, of } from 'rxjs'
import { distinctUntilChanged, map } from 'rxjs/operators'
import { Range, TextDocument, TextDocumentItem } from 'vscode-languageserver-types'
import { ConfigurationCascade } from '../protocol'
import { Selection, URI } from '../types/textDocument'
import { isEqual } from '../util'
import { Extension } from './extension'

/**
 * A description of the environment represented by the CXP client application.
 *
 * This models the state of editor-like tools that display documents, allow selections and scrolling in documents,
 * and support extension configuration.
 *
 * @template X extension type, to support storing additional properties on extensions
 * @template C configuration cascade type
 */
export interface Environment<X extends Extension = Extension, C extends ConfigurationCascade = ConfigurationCascade> {
    /**
     * The root URI of the environment, or null if there is none (which means the extension is unable to access any
     * documents in the environment).
     */
    readonly root: URI | null

    /**
     * The active component (i.e., file view or editor), or null if there is none.
     *
     * TODO(sqs): Support multiple components.
     */
    readonly component: Component | null

    /** The active extensions, or null if there are none. */
    readonly extensions: X[] | null

    /** The configuration cascade. */
    readonly configuration: C
}

/** An empty CXP environment. */
export const EMPTY_ENVIRONMENT: Environment<any, any> = {
    root: null,
    component: null,
    extensions: null,
    configuration: { merged: {} },
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
    readonly environment: Observable<Environment<X, C>> & { readonly value: Environment<X, C> }

    /** The environment's root URI (and changes to it). */
    readonly root: Observable<URI | null>

    /** The environment's active component (and changes to it). */
    readonly component: Observable<Component | null>

    /** The active component's text document (and changes to it). */
    readonly textDocument: Observable<Pick<TextDocument, 'uri' | 'languageId'> | null>

    /** The environment's configuration cascade (and changes to it). */
    readonly configuration: Observable<C>
}

/** An ObservableEnvironment that always represents the empty environment and never emits changes. */
export const EMPTY_OBSERVABLE_ENVIRONMENT: ObservableEnvironment<any, any> = {
    environment: { ...of(EMPTY_ENVIRONMENT), value: EMPTY_ENVIRONMENT } as ObservableEnvironment<
        any,
        any
    >['environment'],
    root: of(null),
    component: of(null),
    textDocument: of(null),
    configuration: of({}),
}

/**
 * Helper function for creating an ObservableEnvironment from the raw environment Observable.
 *
 * @template X extension type
 * @template C configuration cascade type
 */
export function createObservableEnvironment<X extends Extension, C extends ConfigurationCascade>(
    environment: Observable<Environment<X, C>> & { readonly value: Environment<X, C> }
): ObservableEnvironment<X, C> {
    const component = environment.pipe(
        map(({ component }) => component),
        distinctUntilChanged((a, b) => isEqual(a, b))
    )
    return {
        environment,
        root: environment.pipe(
            map(({ root }) => root),
            distinctUntilChanged()
        ),
        component,
        textDocument: component.pipe(
            map(component => (component ? component.document : null)),
            distinctUntilChanged((a, b) => isEqual(a, b))
        ),
        configuration: environment.pipe(
            map(({ configuration }) => configuration),
            distinctUntilChanged((a, b) => isEqual(a, b))
        ),
    }
}
