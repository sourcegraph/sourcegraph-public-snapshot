import * as React from 'react'
import classNames from 'classnames'
import { PanelContainer } from './PanelContainer'

export const SavedSearchesPanel: React.FunctionComponent<{ className?: string }> = ({ className }) => {
    const loadingDisplay = <div>Loading</div>
    const contentDisplay = <div>Content</div>
    const emptyDisplay = <div>Empty</div>

    const actionButtons = <button type="button">Action</button>

    return (
        <PanelContainer
            className={classNames(className, 'saved-searches-panel')}
            title="Recent searches"
            state="populated"
            loadingContent={loadingDisplay}
            populatedContent={contentDisplay}
            emptyContent={emptyDisplay}
            actionButtons={actionButtons}
        />
    )
}
