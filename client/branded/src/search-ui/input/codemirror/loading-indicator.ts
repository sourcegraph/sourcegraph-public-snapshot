import { completionStatus } from '@codemirror/autocomplete'
import type { Extension } from '@codemirror/state'
import { EditorView, ViewPlugin, type ViewUpdate } from '@codemirror/view'

import styles from '../CodeMirrorQueryInput.module.scss'

const DEFAULT_TIMEOUT = 350

/**
 * Creates an extension that renders a autocompletion loading indicator on the
 * right hand side of the editor (like a right hand side gutter).
 *
 * The indicator will be shown if fetching completion results takes longer than
 * 'timeout' milliseconds. The default value is 350ms. Using a value > 0
 * prevents the indicator from being shown for static suggestions (which are
 * fast), which otherwise can be quite distracting.
 */
export function loadingIndicator({ timeout = DEFAULT_TIMEOUT }: { timeout?: number } = {}): Extension {
    return [
        ViewPlugin.fromClass(
            class {
                public dom: HTMLSpanElement
                private timeout: NodeJS.Timeout | null = null

                constructor(public view: EditorView) {
                    const spinner = document.createElement('div')
                    // I'm using a class from a CSS module here because I don't know
                    // whether CodeMirror's CSS-in-JS solution supports keyframes.
                    spinner.className = styles.loadingSpinner

                    // The container is used to add some padding between the
                    // document and the spinner
                    this.dom = document.createElement('div')
                    this.dom.append(spinner)
                    // The element is always takes part in the layout. This makes it
                    // easier to ensure that there is always enough space for the
                    // spinner to be visible in inputs that scroll horizontally
                    this.dom.style.visibility = 'hidden'
                    this.view.contentDOM.after(this.dom)
                }

                public destroy(): void {
                    this.dom.remove()
                }

                public update(update: ViewUpdate): void {
                    const isBusy = update.view.contentDOM.getAttribute('aria-busy') === 'true'
                    const status = completionStatus(update.state)

                    if (status === 'pending' || isBusy) {
                        this.showIndicatorDebounced()
                    } else {
                        this.dom.style.visibility = 'hidden'
                        if (this.timeout !== null) {
                            clearTimeout(this.timeout)
                        }
                    }
                }

                private showIndicatorDebounced(): void {
                    if (this.timeout !== null) {
                        clearTimeout(this.timeout)
                    }
                    this.timeout = setTimeout(() => {
                        this.dom.style.visibility = 'visible'
                    }, timeout)
                }
            },
            {
                provide: plugin =>
                    EditorView.scrollMargins.of(view => {
                        const value = view.plugin(plugin)
                        if (!value) {
                            return null
                        }
                        return { right: value.dom.offsetWidth }
                    }),
            }
        ),
    ]
}
