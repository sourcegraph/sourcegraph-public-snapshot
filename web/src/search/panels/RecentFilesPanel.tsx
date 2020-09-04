import * as React from 'react'
import classNames from 'classnames'
import { PanelContainer } from './PanelContainer'

export const RecentFilesPanel: React.FunctionComponent<{ className?: string }> = ({ className }) => {
    const loadingDisplay = <div>Loading</div>
    const contentDisplay = <div>Content</div>
    const emptyDisplay = <div>Empty</div>

    return (
        <PanelContainer
            className={classNames(className, 'recent-files-panel')}
            title="Recent files"
            state="populated"
            loadingContent={loadingDisplay}
            populatedContent={contentDisplay}
            emptyContent={emptyDisplay}
        />
    )
}
