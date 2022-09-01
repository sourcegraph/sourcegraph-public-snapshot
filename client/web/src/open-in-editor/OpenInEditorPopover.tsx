import React from 'react'

import { mdiClose } from '@mdi/js'

import { Button, Icon } from '@sourcegraph/wildcard'

import { EditorSettings } from './editor-settings'

import styles from './OpenInEditorPopover.module.scss'

export interface OpenInEditorPopoverProps {
    editorSettings?: EditorSettings
    togglePopover: () => void
}

/**
 * A popover that displays a searchable list of revisions (grouped by type) for
 * the current repository.
 */
export const OpenInEditorPopover: React.FunctionComponent<
    React.PropsWithChildren<OpenInEditorPopoverProps>
> = props => {
    const { editorSettings, togglePopover } = props

    return (
        <div className={styles.openInEditorPopover}>
            <Button onClick={togglePopover} variant="icon" className={styles.close} aria-label="Close">
                <Icon aria-hidden={true} svgPath={mdiClose} />
            </Button>
            {editorSettings?.editorId}
        </div>
    )
}
