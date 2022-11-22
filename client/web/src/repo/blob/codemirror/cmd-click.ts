import { Facet, StateEffect, StateField } from '@codemirror/state'
import { EditorView, PluginValue, ViewPlugin } from '@codemirror/view'

const setCmdHeld = StateEffect.define<boolean>()
const isCmdHeld = StateField.define<boolean>({
    create: () => false,
    update: (value, transactions) => {
        for (const effect of transactions.effects) {
            if (effect.is(setCmdHeld)) {
                value = effect.value
            }
        }
        return value
    },
})
const cmdPointerCursor = ViewPlugin.fromClass(
    class implements PluginValue {
        constructor(public view: EditorView) {
            document.body.addEventListener('keydown', this.onKeyDown)
            document.body.addEventListener('keyup', this.onKeyUp)
        }
        public destroy(): void {
            document.body.removeEventListener('keydown', this.onKeyDown)
            document.body.removeEventListener('keyup', this.onKeyUp)
        }

        public onKeyUp = (): void => {
            this.view.contentDOM.classList.remove('cm-token-selection-clickable')
            this.view.dispatch({ effects: setCmdHeld.of(false) })
        }
        public onKeyDown = (event: KeyboardEvent): void => {
            if (event.metaKey) {
                this.view.contentDOM.classList.add('cm-token-selection-clickable')
                this.view.dispatch({ effects: setCmdHeld.of(true) })
            }
        }
    }
)
export const cmdClickFacet = Facet.define<boolean, boolean>({
    combine: sources => sources[0],
    enables: [isCmdHeld, cmdPointerCursor],
})
