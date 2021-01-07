import { SaveToolbar, Props as SaveToolbarProps } from './SaveToolbar'
import * as React from 'react'
import CheckIcon from 'mdi-react/CheckIcon'

export interface AutoIndexProps {
    onQueueJob?: () => void
}

export const CodeIntelAutoIndexSaveToolbar: React.FunctionComponent<SaveToolbarProps & AutoIndexProps> = ({ dirty, saving, error, onSave, onDiscard, onQueueJob }) => {
    return (
        <SaveToolbar
            dirty={dirty}
            saving={saving}
            onSave={onSave}
            onDiscard={onDiscard}>
            <button
                type="button"
                title="BANANA"
                className="btn btn-sm btn-secondary save-toolbar__item save-toolbar__btn save-toolbar__btn-last test-save-toolbar-discard"
                onClick={onQueueJob}>
                <CheckIcon className="icon-inline" /> Bananad
        </button>
        </SaveToolbar>
    )
}
