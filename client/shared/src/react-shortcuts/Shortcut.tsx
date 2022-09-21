import * as React from 'react'

import { Key, ModifierKey } from './keys'
import { Consumer, Context } from './ShortcutProvider'

export interface Props {
    ordered: Key[]
    held?: ModifierKey[]
    node?: HTMLElement | null
    ignoreInput?: boolean
    onMatch(matched: { ordered: Key[]; held?: ModifierKey[] }): void
    allowDefault?: boolean
}

export interface Subscription {
    unsubscribe(): void
}

export const Shortcut: React.FunctionComponent<Props> = props => (
    <Consumer>{(context: Context) => <ShortcutConsumer {...props} {...context} />}</Consumer>
)

class ShortcutConsumer extends React.Component<Props & Context, never> {
    public data = {
        node: this.props.node,
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
