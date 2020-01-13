import * as React from 'react'
import { startCase } from 'lodash'
import { FilterTypes, filterTypeKeys } from '../../../../../shared/src/search/interactive/util'

interface Props {
    onAddNewFilter: (filterType: FilterTypes) => void
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
        this.props.onAddNewFilter(e.target.value as FilterTypes)
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
                    .filter(filter => filter in FilterTypes && filter !== FilterTypes.case)
                    .map(filter => (
                        <option key={filter} value={filter}>
                            {startCase(filter)}
                        </option>
                    ))}
            </select>
        )
    }
}
