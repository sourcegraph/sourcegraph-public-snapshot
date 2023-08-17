import * as React from 'react'

import type { Key, ModifierKey } from './keys'
import type { Data } from './ShortcutManager'
import { Consumer, type Context } from './ShortcutProvider'

export interface Props {
    /**
     * Order in which they keys should be pressed. At the moment the values used
     * need to take into account any modifier key (e.g. to register a keyboard
     * shortcut for 'Alt+t', you'd have to specify the character '†').
     */
    ordered: Key[]
    /**
     * Which modifiers keys need to be pressed for the handler to be triggered.
     * 'Mod' is a special value that translates to 'Meta' (Cmd) on macOS and
     * 'Ctrl' on other platforms.
     */
    held?: (ModifierKey | 'Mod')[]
    /**
     * Be default keybindings are not triggered when the keyboard events
     * originate from an input element. Set this to `true` to also react to
     * keyboard events originaing from input elements.
     */
    ignoreInput?: boolean
    /**
     * The function to be called when the keybinding is triggered.
     */
    onMatch(matched: { ordered: Key[]; held?: ModifierKey[] }): void
    /**
     * By default the browser's default action for the event is prevented. Set
     * to `true` to allow the default action.
     */
    allowDefault?: boolean
}

export interface Subscription {
    unsubscribe(): void
}

/**
 * Registers a global keyboard shortcut.
 */
export const Shortcut: React.FunctionComponent<Props> = props => (
    <Consumer>{(context: Context) => <ShortcutConsumer {...props} {...context} />}</Consumer>
)

class ShortcutConsumer extends React.Component<Props & Context, never> {
    public data: Data = {
        node: null,
        ordered: this.props.ordered,
        held: this.props.held,
        ignoreInput: this.props.ignoreInput || false,
        onMatch: this.props.onMatch,
        allowDefault: this.props.allowDefault || false,
    }
    public subscription: Subscription | null = null

    public componentDidMount(): void {
        const { node } = this.data

        if (node !== null && node !== undefined) {
            return
        }

        const { shortcutManager } = this.props
        if (shortcutManager === undefined) {
            return
        }
        this.subscription = shortcutManager.subscribe(this.data)
    }

    public componentWillUnmount(): void {
        if (this.subscription === null) {
            return
        }

        this.subscription.unsubscribe()
    }

    public render(): React.ReactNode {
        return null
    }
}
