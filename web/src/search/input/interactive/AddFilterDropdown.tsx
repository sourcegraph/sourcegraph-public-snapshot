import * as React from 'react'
import { SuggestionTypes } from '../../../../../shared/src/search/suggestions/util'
import { startCase } from 'lodash'
import { filterTypeKeys, FilterTypes } from './filters'

interface Props {
    onAddNewFilter: (filterType: SuggestionTypes) => void
}

interface State {
    value: string
}

export class AddFilterDropdown extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)

        this.state = { value: 'default' }
    }

    private onAddNewFilter = (e: React.ChangeEvent<HTMLSelectElement>): void => {
        this.props.onAddNewFilter(e.target.value as SuggestionTypes)
        this.setState({ value: 'default' })
    }

    public render(): JSX.Element | null {
        return (
            <select
                className="form-control add-filter-dropdown"
                onChange={this.onAddNewFilter}
                value={this.state.value}
            >
                <option value="default" disabled={true}>
                    Add filter…
                </option>
                {filterTypeKeys
                    .filter(filter => filter in FilterTypes)
                    .map(filter => (
                        <option key={filter} value={filter}>
                            {startCase(filter)}
                        </option>
                    ))}
            </select>
        )
    }
}
