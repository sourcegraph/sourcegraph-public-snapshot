import * as React from 'react'

interface Props {
    /** The user count input by the user. */
    value: number | null

    /** Called when the user count value changes. */
    onChange: (value: number | null) => void

    disabled?: boolean
    className?: string
}

/**
 * Displays a form control for inputting the user count for a product subscription.
 */
export class ProductSubscriptionUserCountFormControl extends React.Component<Props> {
    public render(): JSX.Element | null {
        return (
            <div
                className={`product-subscription-user-count-control form-group align-items-center ${
                    this.props.className || ''
                }`}
            >
                <label
                    htmlFor="product-subscription-user-count-control__userCount"
                    className="mb-0 mr-2 font-weight-bold"
                >
                    Users
                </label>
                <div className="d-flex align-items-center">
                    <input
                        id="product-subscription-user-count-control__userCount"
                        type="number"
                        className="form-control w-auto"
                        min={1}
                        step={1}
                        max={50000}
                        required={true}
                        disabled={this.props.disabled}
                        value={this.props.value === null ? '' : this.props.value}
                        onChange={this.onUserCountChange}
                    />
                </div>
            </div>
        )
    }

    private onUserCountChange: React.ChangeEventHandler<HTMLInputElement> = e => {
        // Check for NaN (which is the value if the user deletes the input's value).
        this.props.onChange(Number.isNaN(e.currentTarget.valueAsNumber) ? null : e.currentTarget.valueAsNumber)
    }
}
