import * as React from 'react'
import { startCase } from 'lodash'
import { SuggestionTypes } from '../../../../../shared/src/search/suggestions/util'

interface RowProps {
    /**
     * Whether we're currently on the search homepage.
     */
    isHomepage: boolean
    /**
     * Callback that adds a new filter to the SelectedFilterRow when one of the buttons are clicked.
     * */
    onAddNewFilter: (filter: SuggestionTypes) => void
}

export enum DefaultFilterTypes {
    repo = 'repo',
    file = 'file',
}

/**
 * The row containing the buttons to add new filters in interactive mode.
 * */
export const AddFilterRow: React.FunctionComponent<RowProps> = ({ isHomepage, onAddNewFilter }) => (
    <div className={`add-filter-row ${isHomepage ? 'add-filter-row--homepage' : ''} e2e-add-filter-row`}>
        {Object.keys(DefaultFilterTypes).map(filterType => (
            <AddFilterButton key={filterType} onAddNewFilter={onAddNewFilter} type={filterType as SuggestionTypes} />
        ))}
    </div>
)

interface AddFilterButtonProps {
    type: SuggestionTypes
    onAddNewFilter: (filter: SuggestionTypes) => void
}

class AddFilterButton extends React.Component<AddFilterButtonProps> {
    private onAddNewFilter = (): void => {
        this.props.onAddNewFilter(this.props.type)
    }

    public render(): JSX.Element | null {
        return (
            <button
                type="button"
                className={`add-filter-row__button btn btn-outline-primary e2e-add-filter-button-${this.props.type}`}
                onClick={this.onAddNewFilter}
            >
                + {startCase(this.props.type)} filter
            </button>
        )
    }
}
