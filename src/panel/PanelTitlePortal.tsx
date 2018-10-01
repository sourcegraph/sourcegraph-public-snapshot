import * as React from 'react'
import { Unsubscribable } from 'rxjs'
import { Panel } from './Panel'

interface Props {
    children: React.ReactFragment
}

/**
 * Used to set the panel title from another component's render method.
 *
 * For example:
 *
 * class MyComponent extends React.Component {
 *     public render(): React.ReactNode {
 *         return (
 *             <div>
 *                 <PanelTitlePortal>
 *                     <MyIcon /> My panel title
 *                 </PanelTitlePortal>
 *                 <h1>My component!</h1>
 *             </div>
 *         )
 *     }
 * }
 *
 * The icon and "My panel title" elements will be rendered as the panel's title.
 */
export class PanelTitlePortal extends React.PureComponent<Props> {
    private subscription: Unsubscribable | undefined

    public componentDidMount(): void {
        this.subscription = Panel.setTitle(this.props.children)
    }

    public componentDidUpdate(prevProps: Props): void {
        if (this.props.children !== prevProps.children) {
            if (this.subscription) {
                this.subscription.unsubscribe()
            }
            this.subscription = Panel.setTitle(this.props.children)
        }
    }

    public componentWillUnmount(): void {
        if (this.subscription) {
            this.subscription.unsubscribe()
            this.subscription = undefined
        }
    }

    public forceUpdate(callBack?: () => void): void {
        super.forceUpdate(callBack)
        Panel.forceUpdate()
    }

    public render(): React.ReactNode {
        return null // the element is rendered in Panel, not here
    }
}
