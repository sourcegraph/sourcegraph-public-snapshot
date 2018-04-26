import * as React from 'react'

/**
 * Descriptor of a radio button element.
 */
export interface RadioButtonNode {
    /**
     * Radio button value.
     */
    id: string | string[] | number

    /**
     * Radio button label.
     */
    label: string

    /**
     * `key` property for radio button wrapper element. If not provided, id.toString() is used instead.
     */
    key?: string | number
}

interface Props {
    /**
     * An additional class name to set on the root element.
     */
    className?: string

    /**
     * List of radio button elements to render.
     */
    nodes: RadioButtonNode[]

    /**
     * Handler for when a radio button is selected.
     */
    onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void

    /**
     * Used to determine if an individual radio button should initially be checked.
     */
    checked?: (n: RadioButtonNode) => boolean
}

/**
 * A row of radio buttons.
 */
export class RadioButtons extends React.PureComponent<Props> {
    public render(): React.ReactFragment {
        return (
            <div className="radio-buttons">
                {this.props.nodes.map(n => (
                    <label key={n.key ? n.key : n.id.toString()} className="radio-buttons__item" title={n.label}>
                        <input
                            className={`radio-buttons__input ${this.props.className || ''}`}
                            name="filter"
                            type="radio"
                            onChange={this.props.onChange}
                            value={n.id}
                            checked={this.props.checked ? this.props.checked(n) : false}
                        />{' '}
                        <div className="radio-buttons__label">{n.label}</div>
                    </label>
                ))}
            </div>
        )
    }
}
