import * as React from 'react'
import { Unsubscribable } from 'rxjs'
import { Panel, PanelItem } from './Panel'

interface Props extends PanelItem {}

/**
 * Used to add items to the panel from other components' render methods.
 *
 * For example:
 *
 * class MyComponent extends React.Component {
 *     public render(): React.ReactNode {
 *         return (
 *             <div>
 *                 <PanelItemPortal position="right" element={
 *                     <MyItem key="my-item" foo="bar" />
 *                 } />
 *                 <h1>My component!</h1>
 *             </div>
 *         )
 *     }
 * }
 *
 * The MyItem component will be rendered in the panel (Panel), not inside MyComponent.
 */
export class PanelItemPortal extends React.PureComponent<Props> {
    private subscription: Unsubscribable | undefined

    public componentDidMount(): void {
        this.subscription = Panel.addItem(this.props)
    }

    public componentWillReceiveProps(props: Props): void {
        if (this.props.element !== props.element || this.props.priority !== props.priority) {
            if (this.subscription) {
                this.subscription.unsubscribe()
            }
            this.subscription = Panel.addItem(props)
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
