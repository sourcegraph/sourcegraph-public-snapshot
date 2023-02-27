import * as React from 'react'

import { ShortcutManager } from './ShortcutManager'

export interface Context {
    shortcutManager?: ShortcutManager
}

export interface Props {
    children?: React.ReactNode
}

export const { Provider, Consumer } = React.createContext<Context>({})

export class ShortcutProvider extends React.Component<Props, never> {
    private shortcutManager = new ShortcutManager()

    public componentDidMount(): void {
        this.shortcutManager.setup()
    }

    public render(): React.ReactNode {
        const context: Context = {
            shortcutManager: this.shortcutManager,
        }

        return <Provider value={context}>{this.props.children}</Provider>
    }
}
