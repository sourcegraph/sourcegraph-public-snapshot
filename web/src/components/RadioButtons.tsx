import * as React from 'react'

/**
 * Descriptor of a radio button element.
 */
interface RadioButtonNode {
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

    /**
     * Radio button tooltip.
     */
    tooltip?: string
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
     * id of the currently selected RadioButtonNode.
     */
    selected?: string
}

/**
 * A row of radio buttons.
 */
export class RadioButtons extends React.PureComponent<Props> {
    public render(): React.ReactFragment {
        return (
            <div className="radio-buttons">
                {this.props.nodes.map(n => (
                    <label key={n.key ? n.key : n.id.toString()} className="radio-buttons__item" title={n.tooltip}>
                        <input
                            className={`radio-buttons__input ${this.props.className || ''}`}
                            name="filter"
                            type="radio"
                            onChange={this.props.onChange}
                            value={n.id}
                            checked={n.id === this.props.selected}
                        />{' '}
                        <small>
                            <div className="radio-buttons__label">{n.label}</div>
                        </small>
                    </label>
                ))}
            </div>
        )
    }
}
