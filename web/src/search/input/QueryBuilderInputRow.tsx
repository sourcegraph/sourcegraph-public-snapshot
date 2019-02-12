import * as React from 'react'

interface Props {
    /** The field title */
    title: string
    /** An example displaying the shortcut for this field. */
    placeholder: string
    /** An optional example for sourcegraph.com that displays the shortcut for this field. */
    dotComPlaceholder?: string
    /** An description of the input field that is displayed below the field. */
    description: string
    /**
     * An appropriate identifier for this field to be used as a suffix for CSS classes and testing IDs.
     * Must be a single or hyphenated word, and unique amongst the other fields in the query builder.
     */
    shortName: string
    /** Handler for when an input field changes. */
    onInputChange: (event: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void
    isSourcegraphDotCom: boolean
}

interface State {
    isFocused: boolean
}

export class QueryBuilderInputRow extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = {
            isFocused: false,
        }
    }

    public render(): JSX.Element | null {
        const placeholder = this.props.isSourcegraphDotCom
            ? this.props.dotComPlaceholder || this.props.placeholder
            : this.props.placeholder
        return (
            <div className="query-builder__row">
                <label className="query-builder__row-label" htmlFor={`query-builder__${this.props.shortName}`}>
                    {this.props.title}:
                </label>
                <div className="query-builder__row-input">
                    <input
                        data-testid={`test-${this.props.shortName}`}
                        id={`query-builder__${this.props.shortName}`}
                        className="form-control query-builder__input"
                        spellCheck={false}
                        autoCapitalize="off"
                        autoComplete="off"
                        placeholder={placeholder}
                        onChange={this.props.onInputChange}
                        onFocus={this.toggleIsFocused}
                        onBlur={this.toggleIsFocused}
                    />
                </div>
                <div className="query-builder__row">
                    <div className="query-builder__row-description">
                        <small>{this.props.description}</small>
                    </div>
                    <div className="query-builder__row-example" />
                </div>
            </div>
        )
    }

    private toggleIsFocused = () => {
        this.setState({ isFocused: !this.state.isFocused })
    }
}
