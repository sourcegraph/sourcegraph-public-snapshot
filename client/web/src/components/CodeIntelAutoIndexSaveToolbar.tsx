import DollyIcon from 'mdi-react/DollyIcon'
import * as React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

import { SaveToolbar, SaveToolbarProps } from './SaveToolbar'

export interface AutoIndexProps {
    enqueueing: boolean
    onQueueJob?: () => void
}

export const CodeIntelAutoIndexSaveToolbar: React.FunctionComponent<SaveToolbarProps & AutoIndexProps> = ({
    dirty,
    saving,
    error,
    onSave,
    onDiscard,
    onQueueJob,
    enqueueing,
    saveDiscardDisabled,
}) => (
    <SaveToolbar
        dirty={dirty}
        saving={saving}
        onSave={onSave}
        error={error}
        saveDiscardDisabled={saveDiscardDisabled}
        onDiscard={onDiscard}
    >
        <button
            type="button"
            title="Enqueue an Index Job"
            disabled={enqueueing}
            className="btn btn-sm btn-secondary save-toolbar__item save-toolbar__btn save-toolbar__btn-last test-save-toolbar-discard"
            onClick={onQueueJob}
        >
            <DollyIcon className="icon-inline" style={{ marginRight: '0.15em' }} /> Queue Job
        </button>
        {enqueueing && (
            <span className="save-toolbar__item save-toolbar__message">
                <LoadingSpinner className="icon-inline" /> Enqueueing...
            </span>
        )}
    </SaveToolbar>
)
