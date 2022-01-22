import classNames from 'classnames'
import React from 'react'

import { TemporarySettingsSchema } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'

type ViewMode = TemporarySettingsSchema['catalog.explorer.viewMode']

interface Props {
    viewMode: ViewMode
    setViewMode: (newValue: ViewMode) => void
}

export const ViewModeToggle: React.FunctionComponent<Props> = ({ viewMode, setViewMode }) => (
    <div className="btn-group" role="group">
        <button
            type="button"
            className={classNames('btn border', viewMode === 'list' ? 'btn-secondary' : 'text-muted')}
            onClick={() => setViewMode('list')}
        >
            List
        </button>
        <button
            type="button"
            className={classNames('btn border', viewMode === 'graph' ? 'btn-secondary' : 'text-muted')}
            onClick={() => setViewMode('graph')}
        >
            Graph
        </button>
    </div>
)

const DEFAULT_VIEW_MODE: ViewMode = 'list'

export const useViewModeTemporarySettings = (): [
    ViewMode,
    (newValue: ViewMode | ((oldValue: ViewMode | undefined) => ViewMode | undefined)) => void
] => {
    const [viewMode, setViewMode] = useTemporarySetting('catalog.explorer.viewMode', 'list')
    return [viewMode || DEFAULT_VIEW_MODE, setViewMode]
}
