import * as React from 'react'

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

/** A toggle switch input component. */
export class Toggle extends React.PureComponent<Props> {
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
                className={`toggle ${this.props.disabled ? 'toggle__disabled' : ''} ${this.props.className || ''}`}
                id={this.props.id}
                title={this.props.title}
                value={this.props.value ? 1 : 0}
                onClick={this.onClick}
                tabIndex={this.props.tabIndex}
            >
                <span
                    className={`toggle__bar ${this.props.value ? 'toggle__bar--active' : ''} ${
                        this.props.disabled ? 'toggle__bar--disabled' : ''
                    }`}
                />
                <span
                    className={`toggle__knob ${this.props.value ? 'toggle__knob--active' : ''} ${
                        this.props.disabled
                            ? this.props.value
                                ? 'toggle__knob--disabled__active'
                                : 'toggle__knob--disabled'
                            : ''
                    }`}
                />
            </button>
        )
    }
}
