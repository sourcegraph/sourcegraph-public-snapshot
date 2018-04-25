import * as React from 'react'
import { Unsubscribable } from 'rxjs'
import { RepoHeader } from './RepoHeader'

interface Props<C extends React.ReactElement<any>> {
    element: C
    position: 'nav' | 'left' | 'right'

    /**
     * Controls the relative order of header action items. The items are laid out from highest priority (at the
     * beginning) to lowest priority (at the end). The default is 0.
     */
    priority?: number
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
    private subscription: Unsubscribable | undefined

    public componentDidMount(): void {
        if (!this.props.element.key) {
            throw new Error('RepoHeaderActionPortal component element must have a key prop')
        }

        this.subscription = RepoHeader.addAction({
            position: this.props.position,
            priority: this.props.priority || 0,
            element: this.props.element,
        })
    }

    public componentWillReceiveProps(props: Props<C>): void {
        if (
            this.props.element !== props.element ||
            this.props.position !== props.position ||
            this.props.priority !== props.priority
        ) {
            if (this.subscription) {
                this.subscription.unsubscribe()
            }
            this.subscription = RepoHeader.addAction({
                position: props.position,
                priority: props.priority || 0,
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
