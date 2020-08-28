import * as React from 'react'
import { PanelContainer } from './PanelContainer'

export const SavedSearchesPanel: React.FunctionComponent<{}> = () => {
    const loadingDisplay = <div>Loading</div>
    const contentDisplay = <div>Content</div>
    const emptyDisplay = <div>Empty</div>

    const actionButtons = <button type="button">Action</button>

    return (
        <PanelContainer
            className="saved-searches-panel"
            title="Recent Searches"
            state="content"
            loadingDisplay={loadingDisplay}
            contentDisplay={contentDisplay}
            emptyDisplay={emptyDisplay}
            actionButtons={actionButtons}
        />
    )
}
