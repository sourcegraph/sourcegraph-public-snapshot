import React from 'react'
import { Subscription } from 'rxjs'
import { ActivationStatus } from './Activation'

// export interface ActivationClickTargetState {
//     activated?: boolean
//     animate?: boolean
// }

// export interface ActivationClickTargetProps {
//     /**
//      * The activation status to update when the element is clicked.
//      */
//     activation?: ActivationStatus

//     /**
//      * The activation step keys to set to true when the element is clicked.
//      */
//     activationKeys: string[]

//     /**
//      * If defined, the component will absorb the click (preventing propagation), pause
//      * (enough time for the activation animation to finish), and then invoke the
//      * specified function after the pause.
//      */
//     pauseAndRetrigger?: () => void
// }

// /**
//  * Wraps another component where a user click should trigger the completion of one or more
//  * activation steps. A user click that results in one or more incomplete steps to be
//  * completed will trigger the activation animation on this component. Clicks that
//  * do not trigger incomplete steps to complete will not trigger any animation.
//  */
// export class ActivationClickTarget extends React.PureComponent<ActivationClickTargetProps, ActivationClickTargetState> {
//     private subscriptions = new Subscription()
//     constructor(props: ActivationClickTargetProps) {
//         super(props)
//         this.state = {}
//     }
//     public componentDidMount(): void {
//         if (this.props.activation) {
//             this.subscriptions.add(
//                 this.props.activation.completed.subscribe(completed => {
//                     if (!completed) {
//                         return
//                     }
//                     // Completed list is available, so now determine whether it's activated by looking at what's completed
//                     let activated = true
//                     for (const k of this.props.activationKeys) {
//                         if (completed[k] === undefined) {
//                             // ignore keys that aren't in completed
//                             continue
//                         }
//                         if (!completed[k]) {
//                             activated = false
//                             break
//                         }
//                     }
//                     this.setState({ activated })
//                 })
//             )
//         }
//     }
//     public componentWillUnmount(): void {
//         this.subscriptions.unsubscribe()
//     }
//     public clicked = (e: React.MouseEvent<HTMLElement, MouseEvent>) => {
//         if (!this.props.activation) {
//             return
//         }
//         if (this.state.activated === undefined) {
//             return
//         }
//         if (this.state.activated) {
//             return
//         }

//         // Activation is occurring
//         const activatePatch: {
//             [key: string]: boolean
//         } = {}
//         for (const k of this.props.activationKeys) {
//             activatePatch[k] = true
//         }
//         this.props.activation.update(activatePatch)
//         if (this.props.pauseAndRetrigger) {
//             e.preventDefault()
//             e.stopPropagation()
//             setTimeout(this.props.pauseAndRetrigger, 1000)
//         }
//         this.setState({ activated: true, animate: !this.state.activated })
//     }
//     public render(): JSX.Element | null {
//         return (
//             <div onClick={this.clicked} className={`first-use-button ${this.state.animate ? 'animate' : ''}`}>
//                 {this.props.children}
//             </div>
//         )
//     }
// }
