import CaretDownIcon from '@sourcegraph/icons/lib/CaretDown'
import * as React from 'react'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'

interface Props {
    /**
     * An additional class name to set on the root element.
     */
    className?: string

    /**
     * The button label.
     */
    children?: React.ReactFragment

    /**
     * The element to display in the popover.
     */
    popoverElement: React.ReactElement<any>

    /**
     * A unique key for the popover, used to ensure that only one popover is visible
     * at a time on the page.
     */
    popoverKey: string

    /**
     * Hide the popover when this prop changes.
     */
    hideOnChange: string
}

interface State {
    /** Whether the popover is open. */
    open: boolean
}

/**
 * A button that toggles the visibility of a popover.
 */
export class PopoverButton extends React.PureComponent<Props, State> {
    private static opens = new Subject<string>()

    public state: State = { open: false }

    private hides = new Subject<void>()

    private subscriptions = new Subscription()

    private rootRef: HTMLElement | null = null
    private popoverRef: HTMLDivElement | null = null

    constructor(props: Props) {
        super(props)

        this.subscriptions.add(() => this.setGlobalListeners(false))
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            PopoverButton.opens.subscribe(popoverKey => {
                if (this.props.popoverKey === popoverKey) {
                    this.setState({ open: true })
                } else if (this.state.open) {
                    // Another popover was opened; close this one.
                    this.setState({ open: false })
                }
            })
        )

        this.subscriptions.add(this.hides.subscribe(() => this.setState({ open: false })))
    }

    public componentWillReceiveProps(props: Props): void {
        if (props.hideOnChange !== this.props.hideOnChange) {
            this.hides.next()
        }
    }

    public componentWillUpdate(props: Props, state: State): void {
        if (state.open !== this.state.open) {
            this.setGlobalListeners(state.open)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private setGlobalListeners(add: boolean): void {
        if (add) {
            window.addEventListener('click', this.onGlobalClick, { capture: true })
            window.addEventListener('keydown', this.onGlobalKeyDown, { capture: true })
        } else {
            window.removeEventListener('click', this.onGlobalClick, { capture: true })
            window.removeEventListener('keydown', this.onGlobalKeyDown, { capture: true })
        }
    }

    private onGlobalClick = (e: MouseEvent): void => {
        if (!this.rootRef || !elementIsDescendent(e.target as HTMLElement, this.rootRef)) {
            // Clicks outside of the popover close it.
            this.hides.next()
        }
    }

    public render(): React.ReactFragment {
        return (
            <div
                className={`popover-button ${this.state.open ? 'popover-button--open' : ''} ${this.props.className ||
                    ''}`}
                onClick={this.onClick}
                ref={this.setRootRef}
            >
                {this.props.children} <CaretDownIcon className="icon-inline popover-button__icon" />
                <div ref={this.setPopoverRef} className="popover-button__popover">
                    {this.state.open && this.props.popoverElement}
                </div>
            </div>
        )
    }

    private onClick = (e: React.MouseEvent<HTMLElement>): void => {
        if (this.state.open) {
            // Clicking within the popover element should not hide.
            if (this.popoverRef && !elementIsDescendent(e.target as HTMLElement, this.popoverRef, this.rootRef)) {
                this.hides.next()
            }
        } else {
            PopoverButton.opens.next(this.props.popoverKey)
        }
    }

    private onGlobalKeyDown = (event: KeyboardEvent) => {
        switch (event.key) {
            case 'Escape': {
                event.preventDefault()
                this.hides.next()
                break
            }
        }
    }

    private setRootRef = (e: HTMLElement | null) => (this.rootRef = e)
    private setPopoverRef = (e: HTMLDivElement | null) => (this.popoverRef = e)
}

function elementIsDescendent(
    candidate: HTMLElement,
    candidateAncestor: HTMLElement,
    boundary?: HTMLElement | null
): boolean {
    let e: HTMLElement | null = candidate
    while (e && e !== boundary) {
        if (e === candidateAncestor) {
            return true
        }
        e = e.parentElement
    }
    return false
}
