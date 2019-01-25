import { Shortcut, ShortcutProps } from '@slimsag/react-shortcuts'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import * as React from 'react'
import Popover, { PopoverProps } from 'reactstrap/lib/Popover'
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

    /**
     * A keybinding  that toggles the visibility of this element.
     */
    toggleVisibilityKeybinding?: Pick<ShortcutProps, 'held' | 'ordered'>[]

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

    private rootRef: HTMLElement | null = null
    private popoverRef: HTMLElement | null = null

    public componentWillReceiveProps(props: Props): void {
        if (props.hideOnChange !== this.props.hideOnChange) {
            this.hide()
        }
    }

    public componentDidMount(): void {
        window.addEventListener('mousedown', this.onClickOutside)
        window.addEventListener('touchstart', this.onClickOutside)
    }

    public componentWillUnmount(): void {
        window.removeEventListener('mousedown', this.onClickOutside)
        window.removeEventListener('touchstart', this.onClickOutside)
    }

    public render(): React.ReactFragment {
        const isOpen = this.state.open || this.props.open

        const popoverAnchor = this.rootRef && (
            <Popover
                placement={this.props.placement || 'auto-start'}
                isOpen={isOpen}
                toggle={this.toggleVisibility}
                target={this.rootRef}
                className={`popover-button2__popover ${this.props.popoverClassName || ''}`}
                // This popover is manually triggered. Must remove default "click" trigger so that
                // in link mode (this.props.link), only caret (not link) opens the popover.
                trigger=""
            >
                {isOpen && <Shortcut ordered={['Escape']} onMatch={this.toggleVisibility} ignoreInput={true} />}
                <div ref={this.setPopoverRef}>{this.props.popoverElement}</div>
            </Popover>
        )

        return (
            <div
                className={`popover-button2 ${isOpen ? 'popover-button2--open' : ''} ${this.props.className || ''} ${
                    this.props.link ? 'popover-button2__container' : 'popover-button2__btn popover-button2__anchor'
                }`}
                ref={this.setRootRef}
            >
                <LinkOrSpan
                    className={
                        this.props.link
                            ? 'popover-button2__btn popover-button2__btn--link'
                            : 'popover-button2__container'
                    }
                    to={this.props.link}
                    onClick={this.props.link ? this.hide : this.toggleVisibility}
                >
                    {this.props.children}{' '}
                    {!this.props.link && <MenuDownIcon className="icon-inline popover-button2__icon" />}
                </LinkOrSpan>
                {this.props.link ? (
                    <div className="popover-button2__anchor">
                        <div onClick={this.toggleVisibility}>
                            <MenuDownIcon className="icon-inline popover-button2__icon popover-button2__icon--outside" />
                        </div>
                        {popoverAnchor}
                    </div>
                ) : (
                    popoverAnchor
                )}
                {this.props.toggleVisibilityKeybinding &&
                    !isOpen &&
                    this.props.toggleVisibilityKeybinding.map((keybinding, i) => (
                        <Shortcut key={i} {...keybinding} onMatch={this.toggleVisibility} />
                    ))}
            </div>
        )
    }

    public onClickOutside = (e: MouseEvent | TouchEvent) => {
        if (
            this.popoverRef &&
            this.rootRef &&
            !this.popoverRef.contains(e.target as HTMLElement) &&
            !this.rootRef.contains(e.target as HTMLElement)
        ) {
            this.hide()
        }
    }

    private hide = () => this.setState({ open: false })

    private setPopoverRef = (e: HTMLElement | null) => (this.popoverRef = e)

    private setRootRef = (e: HTMLElement | null) => (this.rootRef = e)

    private toggleVisibility = () => this.setState(prevState => ({ open: !prevState.open }))
}
