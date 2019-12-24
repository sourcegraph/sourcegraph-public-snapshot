import * as React from 'react'
import { suggestionTypeKeys, SuggestionTypes } from '../../../../../shared/src/search/suggestions/util'
import { startCase } from 'lodash'

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
        console.log('on add new filter')
        this.props.onAddNewFilter(e.target.value as SuggestionTypes)
        this.setState({ value: 'default' })
    }

    public render(): JSX.Element | null {
        return (
            <select onChange={this.onAddNewFilter} value={this.state.value}>
                <option value="default" disabled={true}>
                    Add filter…
                </option>
                {suggestionTypeKeys
                    .filter(filter => filter !== 'repo' && filter !== 'file')
                    .map(filter => (
                        <option key={filter} value={filter}>
                            {startCase(filter)}
                        </option>
                    ))}
            </select>
        )
    }
}
