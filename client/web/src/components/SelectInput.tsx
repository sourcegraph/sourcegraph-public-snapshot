import * as React from 'react'

/**
 * Descriptor of a radio button element.
 */
interface SelectInputNode {
    id: string | string[] | number
    values?: Value[]
    key?: string | number
    tooltip?: string
}

interface Value {
    value: string
    label: string
    args: { [name: string]: string | number | boolean }
}

interface Props {
    /**
     * An additional class name to set on the root element.
     */
    className?: string

    /**
     * List of radio button elements to render.
     */
    nodes: SelectInputNode[]

    /**
     * Handler for when a radio button is selected.
     */
    onChange?: (event: React.ChangeEvent<HTMLSelectElement>) => void

    /**
     * id of the currently selected RadioButtonNode.
     */
    selected?: string
}

/**
 * A row of select inputs.
 */
export class SelectInput extends React.PureComponent<Props> {
    public render(): React.ReactFragment {
        return (
            <div>
                {this.props.nodes.map(node => (
                    <select
                        name={node.id.toString()}
                        key={node.key ? node.key : node.id.toString()}
                        onChange={this.props.onChange}
                        className="select-input"
                        value={this.props.selected}
                    >
                        {node.values !== undefined &&
                            node.values!.map(value => (
                                <option key={value.value} className="radio-buttons__item" value={value.value}>
                                    {value.label}
                                </option>
                            ))}
                    </select>
                ))}
            </div>
        )
    }
}
