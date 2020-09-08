import * as React from 'react'
import classNames from 'classnames'
import { PanelContainer } from './PanelContainer'

export const RecentSearchesPanel: React.FunctionComponent<{ className?: string }> = ({ className }) => {
    const loadingDisplay = <div>Loading</div>
    const contentDisplay = <div>Content</div>
    const emptyDisplay = <div>Empty</div>

    return (
        <PanelContainer
            className={classNames(className, 'recent-searches-panel')}
            title="Recent searches"
            state="loading"
            loadingContent={loadingDisplay}
            populatedContent={contentDisplay}
            emptyContent={emptyDisplay}
        />
    )
}
