import * as React from 'react'
import ShortcutManager from '../ShortcutManager'

export interface Context {
    shortcutManager?: ShortcutManager
}

export interface Props {
    children?: React.ReactNode
}

export const { Provider, Consumer } = React.createContext<Context>({})

export default class ShortcutProvider extends React.Component<Props, never> {
    private shortcutManager = new ShortcutManager()

    componentDidMount() {
        this.shortcutManager.setup()
    }

    render() {
        const context: Context = {
            shortcutManager: this.shortcutManager,
        }

        return <Provider value={context}>{this.props.children}</Provider>
    }
}
