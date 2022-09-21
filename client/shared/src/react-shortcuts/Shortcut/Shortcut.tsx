import * as React from 'react'
import { Consumer, Context } from '../ShortcutProvider'
import Key, { ModifierKey } from '../keys'

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

export default function Shortcut(props: Props) {
    return <Consumer>{(context: Context) => <ShortcutConsumer {...props} {...context} />}</Consumer>
}

class ShortcutConsumer extends React.Component<Props & Context, never> {
    public data = {
        node: this.props.node,
        ordered: this.props.ordered,
        held: this.props.held,
        ignoreInput: this.props.ignoreInput || false,
        onMatch: this.props.onMatch,
        allowDefault: this.props.allowDefault || false,
    }
    public subscription!: Subscription

    componentDidMount() {
        const { node } = this.data

        if (node != null) {
            return
        }

        const { shortcutManager } = this.props
        if (shortcutManager == null) {
            return
        }
        this.subscription = shortcutManager.subscribe(this.data)
    }

    componentWillUnmount() {
        if (this.subscription == null) {
            return
        }

        this.subscription.unsubscribe()
    }

    render() {
        return null
    }
}
