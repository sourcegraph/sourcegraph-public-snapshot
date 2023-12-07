import { Compartment, type StateEffect, type Extension } from '@codemirror/state'
import type { EditorView } from '@codemirror/view'

type Extensions = Record<string, Extension>
type UpdatedExtensions<T extends Extensions> = { [key in keyof T]: Extension }

interface Compartments<T extends Extensions> {
    extension: Extension
    /**
     * Initialize compartments with a different value
     */
    init(extensions: UpdatedExtensions<T>): Extension
    /**
     * Update compartments. Only compartments for which the provided
     * value has changed will be updated.
     */
    update(view: EditorView, extensions: UpdatedExtensions<T>): void
}

/**
 * Helper function for creating a compartments extension. Each record
 * entry will get its own compartment and the value will be the initial
 * value of the compartment.
 */
export function createCompartments<T extends Record<string, Extension>>(extensions: T): Compartments<T> {
    const compartments: Record<string, Compartment> = {}

    function init(extensions: UpdatedExtensions<T>, compartments: Record<string, Compartment>): Extension {
        const extension: Extension[] = []
        for (const [name, ext] of Object.entries(extensions)) {
            let compartment = compartments[name]
            if (!compartment) {
                compartment = compartments[name] = new Compartment()
            }
            extension.push(compartment.of(ext))
        }
        return extension
    }

    return {
        extension: init(extensions, compartments),
        init(extensions) {
            return init(extensions, compartments)
        },
        update(view, extensions) {
            const effects: StateEffect<unknown>[] = []

            for (const [name, ext] of Object.entries(extensions)) {
                if (compartments[name].get(view.state) !== ext) {
                    effects.push(compartments[name].reconfigure(ext))
                }
            }

            if (effects.length > 0) {
                view.dispatch({ effects })
            }
        },
    }
}
