import { Compartment, type StateEffect, type Extension } from '@codemirror/state'
import { EditorView } from '@codemirror/view'

type Extensions = Record<string, Extension | null>
type UpdatedExtensions<T extends Extensions> = { [key in keyof T]?: Extension | null }
export type ExtensionType<T> = T extends Compartments<infer U> ? UpdatedExtensions<U> : never

interface Compartments<T extends Extensions> {
    extension: Extension
    /**
     * Initialize compartments with a different value.
     *
     * @param extensions The values for the compartments
     * @returns An extension
     */
    init(extensions: UpdatedExtensions<T>): Extension
    /**
     * Update compartments. Only compartments for which the provided value has changed will be updated.
     * Additional effects can be provided to be dispatched with the compartment updates, but they will
     * only be dispatched if at least one compartment has changed.
     *
     * @param view The editor view
     * @param extensions The updated values for the compartments
     * @param additionalEffects Additional effects to be dispatched with the compartment updates
     */
    update(view: EditorView, extensions: UpdatedExtensions<T>, ...additionalEffects: StateEffect<unknown>[]): void
}

const emptyExtension: Extension = []

/**
 * Helper function for creating a compartments extension. Each record
 * entry will get its own compartment and the value will be the initial
 * value of the compartment.
 * The order/presedence of the extensions is determined by the order of the
 * keys in the initialExtensions record.
 *
 * @param initialExtensions Initial values for the compartments
 * @returns Compartments extension
 */
export function createCompartments<T extends Extensions>(initialExtensions: T): Compartments<T> {
    const compartments: Map<string, Compartment> = new Map()

    function init(extensions: UpdatedExtensions<T>, compartments: Map<string, Compartment>): Extension {
        const values: Map<string, Extension> = new Map()

        for (const [name, ext] of Object.entries(extensions)) {
            let compartment = compartments.get(name)
            if (!compartment) {
                compartments.set(name, (compartment = new Compartment()))
            }
            values.set(name, compartment.of(ext ?? emptyExtension))
        }

        // Return extensions in the order of the initialExtensions record
        return Array.from(compartments.keys(), name => values.get(name) ?? emptyExtension)
    }

    return {
        extension: init(initialExtensions, compartments),
        init(extensions) {
            return init({ ...initialExtensions, ...extensions }, compartments)
        },
        update(view, extensions, ...additionalEffects) {
            const effects: StateEffect<unknown>[] = []

            for (const [name, ext] of Object.entries(extensions)) {
                const compartment = compartments.get(name)
                if (compartment && compartment.get(view.state) !== ext) {
                    effects.push(compartment.reconfigure(ext ?? emptyExtension))
                }
            }

            if (effects.length > 0) {
                view.dispatch({ effects: effects.concat(additionalEffects) })
            }
        },
    }
}

/**
 * An object representing the scroll state of a CodeMirror instance.
 */
export interface ScrollSnapshot {
    scrollTop?: number
}

/**
 * Returns a snapshot of the editors scroll state that is serializable (unlike CodeMirror's
 * native scroll snapshot).
 *
 * @param view The editor view
 * @returns The scroll snapshot
 */
export function getScrollSnapshot(view: EditorView): ScrollSnapshot {
    return {
        scrollTop: view.scrollDOM.scrollTop,
    }
}

/**
 * Restores the scroll state of the editor from a snapshot.
 *
 * @param view The editor view
 * @param snapshot The scroll snapshot
 */
export function restoreScrollSnapshot(view: EditorView, snapshot: ScrollSnapshot): void {
    const { scrollTop } = snapshot

    if (scrollTop !== undefined) {
        // We are using request measure here to ensure that the DOM has been updated
        // before updating the scroll position.
        view.requestMeasure({
            read() {
                return null
            },
            write(_measure, view) {
                view.scrollDOM.scrollTop = scrollTop
            },
        })
    }
}
