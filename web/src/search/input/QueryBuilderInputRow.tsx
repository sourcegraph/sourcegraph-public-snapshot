import * as React from 'react'
import { QueryBuilderState } from './QueryBuilder'

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
     * An appropriate identifier for this field, to be used as a suffix for CSS classes and testing IDs.
     * Must be a single or hyphenated word, and unique amongst the other fields in the query builder.
     */
    shortName: keyof QueryBuilderState['fields']
    /** Handler for when an input field changes. */
    onInputChange: (
        key: keyof QueryBuilderState['fields']
    ) => (event: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void
    isSourcegraphDotCom: boolean
}

export const QueryBuilderInputRow: React.FunctionComponent<Props> = props => {
    const placeholder = props.isSourcegraphDotCom ? props.dotComPlaceholder || props.placeholder : props.placeholder
    return (
        <div className="query-builder-input-row">
            <label className="query-builder-input-row__label" htmlFor={`query-builder__${props.shortName}`}>
                {props.title}:
            </label>
            <div className="query-builder-input-row__input">
                <input
                    data-testid={`test-${props.shortName}`}
                    id={`query-builder-${props.shortName}`}
                    className="form-control query-builder__input"
                    spellCheck={false}
                    autoCapitalize="off"
                    autoComplete="off"
                    placeholder={placeholder}
                    onChange={props.onInputChange(props.shortName)}
                />
            </div>
            <div className="query-builder-input-row">
                <div className="query-builder-input-row__description">
                    <small>{props.description}</small>
                </div>
            </div>
        </div>
    )
}
