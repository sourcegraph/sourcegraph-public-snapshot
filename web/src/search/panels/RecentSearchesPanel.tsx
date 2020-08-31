import * as React from 'react'
import { PanelContainer } from './PanelContainer'

export const RecentSearchesPanel: React.FunctionComponent<{}> = () => {
    const loadingDisplay = <div>Loading</div>
    const contentDisplay = <div>Content</div>
    const emptyDisplay = <div>Empty</div>

    return (
        <PanelContainer
            className="recent-searches-panel"
            title="Recent searches"
            state="loading"
            loadingContent={loadingDisplay}
            populatedContent={contentDisplay}
            emptyContent={emptyDisplay}
        />
    )
}
