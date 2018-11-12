import PinIcon from 'mdi-react/PinIcon'
import * as React from 'react'
import { SavedQueryRow } from './SavedQueryRow'

export interface IExampleSearch {
    query: string
    description: string
}

interface Props {
    search: IExampleSearch
    onSave: (q: IExampleSearch) => void
    isLightTheme: boolean
    isHidden: boolean
}

export const ExampleSearch = (props: Props) => {
    const handleSave = (e: React.MouseEvent<HTMLButtonElement>) => {
        e.stopPropagation()
        e.preventDefault()

        props.onSave(props.search)
    }

    return (
        <SavedQueryRow
            query={props.search.query}
            description={props.search.description}
            eventName="ExampleSearchClick"
            isLightTheme={props.isLightTheme}
            className={props.isHidden ? 'example-searches__hidden' : ''}
            actions={
                <div className="saved-query-row__actions">
                    <button className="btn btn-icon action" onClick={handleSave}>
                        <PinIcon className="icon-inline" />
                        Save
                    </button>
                </div>
            }
        />
    )
}
