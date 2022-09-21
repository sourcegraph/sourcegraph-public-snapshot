import { Key, ModifierKey } from './keys'

const ON_MATCH_DELAY = 500

export interface Data {
    node: HTMLElement | null | undefined
    ordered: Key[]
    held?: ModifierKey[]
    ignoreInput: boolean
    onMatch(matched: { ordered: Key[]; held?: ModifierKey[] }): void
    allowDefault: boolean
}

export class ShortcutManager {
    private keysPressed: Key[] = []
    private shortcuts: Data[] = []
    private shortcutsMatched: Data[] = []
    private timer!: number

    public setup(): void {
        document.addEventListener('keydown', this.handleKeyDown)
    }

    public subscribe(data: Data): { unsubscribe: () => void } {
        const { shortcuts } = this

        shortcuts.push(data)

        return {
            unsubscribe() {
                const unsubscribeIndex = shortcuts.findIndex(shortcut => shortcut === data)
                shortcuts.splice(unsubscribeIndex, 1)
            },
        }
    }

    private resetKeys(): void {
        this.keysPressed = []
        this.shortcutsMatched = []
    }

    private handleKeyDown = (event: KeyboardEvent): void => {
        const { key } = event

        this.keysPressed.push(key as Key)
        this.updateMatchingShortcuts(event)

        switch (this.shortcutsMatched.length) {
            case 0:
                this.resetKeys()
                break
            case 1:
                this.callMatchedShortcut(event)
                break
            default:
                this.timer = window.setTimeout(() => {
                    this.callMatchedShortcut(event)
                }, ON_MATCH_DELAY)
        }
    }

    private updateMatchingShortcuts(event: KeyboardEvent): void {
        const shortcuts = this.shortcutsMatched.length > 0 ? this.shortcutsMatched : this.shortcuts

        this.shortcutsMatched = shortcuts.filter(({ ordered, held, node, ignoreInput }) => {
            if (isFocusedInput() && !ignoreInput) {
                return false
            }

            if (held && !held.every(key => event.getModifierState(key))) {
                return false
            }

            const partiallyMatching = arraysMatch(this.keysPressed, ordered.slice(0, this.keysPressed.length))

            if (node) {
                const onFocusedNode = document.activeElement === node
                return partiallyMatching && onFocusedNode
            }

            return partiallyMatching
        })
    }

    private callMatchedShortcut(event: Event): void {
        const longestMatchingShortcut = this.shortcutsMatched.find(({ ordered }) =>
            arraysMatch(ordered, this.keysPressed)
        )

        if (!longestMatchingShortcut) {
            return
        }

        if (!longestMatchingShortcut.allowDefault) {
            event.preventDefault()
        }

        longestMatchingShortcut.onMatch({
            ordered: longestMatchingShortcut.ordered,
            held: longestMatchingShortcut.held,
        })

        clearTimeout(this.timer)

        this.resetKeys()
    }
}

function isFocusedInput(): boolean {
    const target = document.activeElement
    if (target?.tagName === null) {
        return false
    }

    return Boolean(
        target?.tagName === 'INPUT' ||
            target?.tagName === 'SELECT' ||
            target?.tagName === 'TEXTAREA' ||
            target?.hasAttribute('contenteditable')
    )
}

function arraysMatch<T>(first: T[], second: T[]): boolean {
    if (first.length !== second.length) {
        return false
    }

    return first.every((value, index) => second[index] === value)
}
