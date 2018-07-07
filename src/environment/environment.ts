import { Observable, of } from 'rxjs'
import { distinctUntilChanged, map } from 'rxjs/operators'
import { Range, TextDocument } from 'vscode-languageserver-types'
import { Selection, URI } from '../types/textDocument'
import { isEqual } from '../util'
import { Extension } from './extension'

/**
 * A description of the environment represented by the CXP client application.
 *
 * This models the state of editor-like tools that display documents, allow selections and scrolling in documents,
 * and support extension configuration.
 */
export interface Environment {
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
    readonly extensions: Extension[] | null
}

/** An empty CXP environment. */
export const EMPTY_ENVIRONMENT: Environment = { root: null, component: null, extensions: null }

/** An application component that displays a [TextDocument](#TextDocument). */
export interface Component {
    /** The document displayed by the component. */
    readonly document: Pick<TextDocument, 'uri' | 'languageId'>

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
 */
export interface ObservableEnvironment {
    /** The environment (and changes to it). */
    readonly environment: Observable<Environment> & { readonly value: Environment }

    /** The environment's root URI (and changes to it). */
    readonly root: Observable<URI | null>

    /** The environment's active component (and changes to it). */
    readonly component: Observable<Component | null>

    /** The active component's text document (and changes to it). */
    readonly textDocument: Observable<TextDocument | null>
}

/** An ObservableEnvironment that always represents the empty environment and never emits changes. */
export const EMPTY_OBSERVABLE_ENVIRONMENT: ObservableEnvironment = {
    environment: { ...of(EMPTY_ENVIRONMENT), value: EMPTY_ENVIRONMENT } as ObservableEnvironment['environment'],
    root: of(null),
    component: of(null),
    textDocument: of(null),
}

export function createObservableEnvironment(
    environment: Observable<Environment> & { readonly value: Environment }
): ObservableEnvironment {
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
    }
}
