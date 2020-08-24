import * as React from 'react'
import Close from 'mdi-react/CloseIcon'
import Check from 'mdi-react/CheckIcon'

interface Props {
    /** The initial value. */
    value?: boolean

    /** The DOM ID of the element. */
    id?: string

    /**
     * Called when the user changes the input's value.
     */
    onToggle?: (value: boolean) => void

    /**
     * Called when the user clicks the toggle when it is disabled.
     */
    onToggleDisabled?: (value: boolean) => void

    /** The title attribute (tooltip). */
    title?: string

    disabled?: boolean
    tabIndex?: number
    className?: string
}

/** A toggle switch input component.
 *
 * TODO: Make it big
 */
export class ToggleBig extends React.PureComponent<Props> {
    private onClick = (): void => {
        if (!this.props.disabled && this.props.onToggle) {
            this.props.onToggle(!this.props.value)
        } else if (this.props.disabled && this.props.onToggleDisabled) {
            this.props.onToggleDisabled(!!this.props.value)
        }
    }
    public render(): JSX.Element | null {
        return (
            <button
                type="button"
                className={`toggle-big ${this.props.disabled ? 'toggle-big__disabled' : ''} ${
                    this.props.className || ''
                }`}
                id={this.props.id}
                title={this.props.title}
                value={this.props.value ? 1 : 0}
                onClick={this.onClick}
                tabIndex={this.props.tabIndex}
            >
                <span className="toggle-big__container">
                    <span
                        className={`toggle-big__bar ${this.props.value ? 'toggle-big__bar--active' : ''} ${
                            this.props.disabled ? 'toggle-big__bar--disabled' : ''
                        }`}
                    />
                    <span className={`toggle-big__knob ${this.props.value ? 'toggle-big__knob--active' : ''}`} />
                    <span className={`toggle-big__text ${!this.props.value ? 'toggle-big__text--disabled' : ''}`}>
                        {this.props.value ? 'Enabled' : 'Disabled'}
                    </span>
                    {this.props.value ? (
                        <Check size={16} className="toggle-big__icon" />
                    ) : (
                        <Close size={16} className="toggle-big__icon toggle-big__icon--disabled" />
                    )}
                </span>
            </button>
        )
    }
}
