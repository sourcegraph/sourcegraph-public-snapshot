import * as React from 'react'
import { AnonymousSubscription } from 'rxjs/Subscription'
import { RepoHeader } from './RepoHeader'

/**
 * An action link that is added to and displayed in the repository header.
 */
export interface RepoHeaderAction {
    position: 'left' | 'right'
    component: React.ReactElement<any>
}

/**
 * Used to add actions to the repository header from other components' render methods.
 *
 * For example:
 *
 * class MyComponent extends React.Component {
 *     public render(): React.ReactNode {
 *         return (
 *             <div>
 *                 <RepoHeaderActionPortal position="right" component={
 *                     <MyAction key="toggle-rendered-file-mode" foo="bar" />
 *                 } />
 *                 <h1>My component!</h1>
 *             </div>
 *         )
 *     }
 * }
 *
 * The MyAction component will be rendered in the repository header (RepoHeader), not
 * inside MyComponent.
 *
 * See design note in the RepoHeader docstring.
 */
export class RepoHeaderActionPortal<C extends React.ReactElement<any>> extends React.PureComponent<{
    component: C
    position: 'left' | 'right'
}> {
    private subscription: AnonymousSubscription | undefined

    public componentDidMount(): void {
        if (!this.props.component.key) {
            throw new Error('RepoHeaderActionPortal component element must have a key prop')
        }

        this.subscription = RepoHeader.addAction({
            position: this.props.position,
            component: this.props.component,
        })
    }

    public componentWillReceiveProps(props: { component: C; position: 'left' | 'right' }): void {
        if (this.props.component !== props.component || this.props.position !== props.position) {
            if (this.subscription) {
                this.subscription.unsubscribe()
            }
            this.subscription = RepoHeader.addAction({
                position: props.position,
                component: props.component,
            })
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
        RepoHeader.forceUpdate()
    }

    public render(): React.ReactNode {
        return null // the element is rendered in RepoHeader, not here
    }
}
