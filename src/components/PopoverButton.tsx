import MenuDownIcon from 'mdi-react/MenuDownIcon'
import * as React from 'react'
import Popover, { PopoverProps } from 'reactstrap/lib/Popover'
import { Subscription } from 'rxjs'
import { Key } from 'ts-key-enum'
import { LinkOrSpan } from './LinkOrSpan'

interface Props {
    /**
     * An additional class name to set on the root element.
     */
    className?: string

    /** An additional class name for the popover element. */
    popoverClassName?: string

    /**
     * The button label.
     */
    children?: React.ReactFragment

    /**
     * The link destination URL for the button. If set, the caret is outside of the button.
     */
    link?: string

    /**
     * The element to display in the popover.
     */
    popoverElement: React.ReactElement<any>

    /**
     * Hide the popover when this prop changes.
     */
    hideOnChange?: any

    /** If set, pressing this key toggles the popover's open state. */
    globalKeyBinding?: string

    /**
     * Whether the global keybinding should be active even when the user has an input element focused. This should
     * only be used for keybindings that would not conflict with routine user input. For example, use it for a
     * keybinding to F1, which will ensure that F1 doesn't open Chrome help when the user is focused in an input
     * field.
     *
     */
    globalKeyBindingActiveInInputs?: boolean

    /** Popover placement. */
    placement?: PopoverProps['placement']

    /** Force open, if true. */
    open?: boolean
}

interface State {
    /** Whether the popover is open. */
    open: boolean
}
/**
 * A button that toggles the visibility of a popover.
 */
export class PopoverButton extends React.PureComponent<Props, State> {
    public state: State = { open: false }

    private subscriptions = new Subscription()

    private rootRef: HTMLElement | null = null

    public componentDidMount(): void {
        window.addEventListener('keydown', this.onGlobalKeyDown)
    }

    public componentWillReceiveProps(props: Props): void {
        if (props.hideOnChange !== this.props.hideOnChange) {
            this.setState({ open: false })
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
        window.removeEventListener('keydown', this.onGlobalKeyDown)
    }

    public render(): React.ReactFragment {
        const isOpen = this.state.open || this.props.open
        const popoverAnchor = this.rootRef && (
            <Popover
                placement={this.props.placement || 'auto-start'}
                isOpen={isOpen}
                toggle={this.onPopoverVisibilityToggle}
                target={this.rootRef}
                className={`popover-button__popover ${this.props.popoverClassName || ''}`}
            >
                {this.props.popoverElement}
            </Popover>
        )
        return (
            <div
                className={`popover-button ${isOpen ? 'popover-button--open' : ''} ${this.props.className || ''} ${
                    this.props.link ? 'popover-button__container' : 'popover-button__btn popover-button__anchor'
                }`}
                ref={this.setRootRef}
            >
                <LinkOrSpan
                    className={
                        this.props.link ? 'popover-button__btn popover-button__btn--link' : 'popover-button__container'
                    }
                    to={this.props.link}
                    onClick={this.props.link ? this.onClickLink : this.onPopoverVisibilityToggle}
                >
                    {this.props.children}{' '}
                    {!this.props.link && <MenuDownIcon className="icon-inline popover-button__icon" />}
                </LinkOrSpan>
                {this.props.link ? (
                    <button className="popover-button__anchor btn-icon" onClick={this.onPopoverVisibilityToggle}>
                        <MenuDownIcon className="icon-inline popover-button__icon popover-button__icon--outside" />
                        {popoverAnchor}
                    </button>
                ) : (
                    popoverAnchor
                )}
            </div>
        )
    }

    private onClickLink = (e: React.MouseEvent<HTMLElement>): void => {
        this.setState({ open: false })
    }

    private onGlobalKeyDown = (event: KeyboardEvent) => {
        if (event.key === Key.Escape) {
            // Always close the popover when Escape is pressed, even when in an input.
            this.setState({ open: false })
            return
        }

        // Otherwise don't interfere with user keyboard input.
        if (!this.props.globalKeyBindingActiveInInputs && isInputLike(event.target as HTMLElement)) {
            return
        }

        if (
            this.props.globalKeyBinding &&
            !event.ctrlKey &&
            !event.altKey &&
            !event.metaKey &&
            event.key === this.props.globalKeyBinding
        ) {
            event.preventDefault()
            this.onPopoverVisibilityToggle()
        }
    }

    private setRootRef = (e: HTMLElement | null) => (this.rootRef = e)

    private onPopoverVisibilityToggle = () => this.setState(prevState => ({ open: !prevState.open }))
}

/** Reports whether elem is a field that accepts user keyboard input.  */
function isInputLike(elem: HTMLElement): boolean {
    return elem.tagName === 'INPUT' || elem.tagName === 'TEXTAREA' || elem.tagName === 'SELECT'
}
