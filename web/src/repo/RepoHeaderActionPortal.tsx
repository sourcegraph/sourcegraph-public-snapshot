import * as React from 'react'
import { AnonymousSubscription } from 'rxjs/Subscription'
import { RepoHeader } from './RepoHeader'

interface Props<C extends React.ReactElement<any>> {
    element: C
    position: 'path' | 'left' | 'right'
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
 *                 <RepoHeaderActionPortal position="right" element={
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
export class RepoHeaderActionPortal<C extends React.ReactElement<any>> extends React.PureComponent<Props<C>> {
    private subscription: AnonymousSubscription | undefined

    public componentDidMount(): void {
        if (!this.props.element.key) {
            throw new Error('RepoHeaderActionPortal component element must have a key prop')
        }

        this.subscription = RepoHeader.addAction({
            position: this.props.position,
            element: this.props.element,
        })
    }

    public componentWillReceiveProps(props: Props<C>): void {
        if (this.props.element !== props.element || this.props.position !== props.position) {
            if (this.subscription) {
                this.subscription.unsubscribe()
            }
            this.subscription = RepoHeader.addAction({
                position: props.position,
                element: props.element,
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
