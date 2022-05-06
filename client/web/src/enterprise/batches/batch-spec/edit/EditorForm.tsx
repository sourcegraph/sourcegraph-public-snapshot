import React, { useContext } from 'react'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Panel } from '@sourcegraph/wildcard'

import { BatchSpecContext } from '../BatchSpecContext'

import { EditorFeedbackPanel } from './editor/EditorFeedbackPanel'
import { MonacoBatchSpecEditor } from './editor/MonacoBatchSpecEditor'
import { LibraryPane } from './library/LibraryPane'
import { WorkspacesPreview } from './workspaces-preview/WorkspacesPreview'

import styles from './EditorForm.module.scss'

const WORKSPACES_PREVIEW_SIZE = 'batch-changes.ssbc-workspaces-preview-size'

interface EditorFormProps extends ThemeProps {}

export const EditorForm: React.FunctionComponent<React.PropsWithChildren<EditorFormProps>> = ({ isLightTheme }) => {
    const { batchChange, editor, errors } = useContext(BatchSpecContext)

    return (
        <div className={styles.form}>
            <LibraryPane name={batchChange.name} onReplaceItem={editor.handleCodeChange} />
            <div className={styles.editorContainer}>
                <h4 className={styles.header}>Batch spec</h4>

                <MonacoBatchSpecEditor
                    batchChangeName={batchChange.name}
                    className={styles.editor}
                    isLightTheme={isLightTheme}
                    value={editor.code}
                    onChange={editor.handleCodeChange}
                />
                <EditorFeedbackPanel errors={errors} />
            </div>
            <Panel
                className="d-flex"
                defaultSize={500}
                minSize={405}
                maxSize={1400}
                position="right"
                storageKey={WORKSPACES_PREVIEW_SIZE}
            >
                <div className={styles.workspacesPreviewContainer}>
                    <WorkspacesPreview />
                </div>
            </Panel>
        </div>
    )
}
