import * as React from 'react'

interface Props {
    /** The initial value.  */
    value?: boolean

    /**
     * Called when the user changes the input's value.
     */
    onToggle?: (value: boolean) => void

    /** The title attribute (tooltip). */
    title: string

    disabled?: boolean
    tabIndex?: number
    className?: string
}

interface State {
    value: boolean | undefined
}

/** A toggle switch input component. */
export class Toggle extends React.PureComponent<Props, State> {
    public state: State = { value: undefined }

    public render(): JSX.Element | null {
        const value = this.state.value === undefined ? this.props.value : this.state.value
        // console.log({ value, stateValue: this.state.value, propsValue: this.props.value })
        return (
            <input
                className={`toggle toggle--${value ? 'on' : 'off'}`}
                title={this.props.title}
                type="range"
                value={value ? 1 : 0}
                onMouseDown={this.onMouseDown}
                onInput={this.onChange} // fires in some cases where onChange doesn't
                onChange={this.onChange}
                onMouseUp={this.onMouseUp}
                disabled={this.props.disabled}
                tabIndex={this.props.tabIndex}
                min={0}
                max={1}
                step={1}
            />
        )
    }

    // Track mousedown state to avoid a click triggering both the change and click events (and toggling the value
    // twice). Also track whether the value changed during mousedown so that clicking and releasing on one side
    // will toggle it.
    private mouseDown = false
    private changedDuringMouseDown = false
    private onMouseDown: React.MouseEventHandler<HTMLInputElement> = e => {
        this.mouseDown = true
        this.changedDuringMouseDown = false
    }

    private onChange: React.FormEventHandler<HTMLInputElement> = e => {
        const value = e.currentTarget.valueAsNumber === 1
        if (this.mouseDown) {
            this.changedDuringMouseDown = true
            this.setState({ value })
        } else {
            this.onToggle(value)
        }
    }

    private onMouseUp: React.MouseEventHandler<HTMLInputElement> = e => {
        if (!this.mouseDown) {
            return
        }
        this.mouseDown = false

        // Clicking and releasing entirely on one side will toggle the current value.
        let value = e.currentTarget.valueAsNumber === 1
        if (!this.changedDuringMouseDown) {
            const rect = e.currentTarget.getBoundingClientRect()
            const mouseOverElement =
                rect.left <= e.pageX &&
                rect.left + rect.width >= e.pageX &&
                rect.top <= e.pageY &&
                rect.top + rect.height >= e.pageY
            if (!mouseOverElement) {
                return
            }
            value = !value
        }
        this.setState({ value: undefined }, () => this.onToggle(value))
    }

    private onToggle(value: boolean): void {
        if (value !== !!this.props.value) {
            if (this.props.onToggle) {
                this.props.onToggle(value)
            }
        }
    }
}
