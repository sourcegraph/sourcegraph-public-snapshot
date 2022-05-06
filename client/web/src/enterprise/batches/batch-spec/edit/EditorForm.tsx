import React, { useContext } from 'react'

import { ApolloQueryResult } from '@apollo/client'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Panel } from '@sourcegraph/wildcard'

import { GetBatchChangeToEditResult } from '../../../../graphql-operations'
import { EditorFeedbackPanel } from '../../create/editor/EditorFeedbackPanel'
import { MonacoBatchSpecEditor } from '../../create/editor/MonacoBatchSpecEditor'
import { BatchSpecContext } from '../BatchSpecContext'

import { LibraryPane } from './library/LibraryPane'
import { WorkspacesPreview } from './workspaces-preview/WorkspacesPreview'

import styles from './EditorForm.module.scss'

const WORKSPACES_PREVIEW_SIZE = 'batch-changes.ssbc-workspaces-preview-size'

interface EditorFormProps extends ThemeProps {
    refetchBatchChange: () => Promise<ApolloQueryResult<GetBatchChangeToEditResult>>
}

export const EditorForm: React.FunctionComponent<React.PropsWithChildren<EditorFormProps>> = ({
    isLightTheme,
    refetchBatchChange,
}) => {
    const { batchChange, batchSpec, editor, errors } = useContext(BatchSpecContext)

    console.log({ errors })

    return (
        <div className={styles.form}>
            <LibraryPane name={batchChange.name} onReplaceItem={() => alert('hi')} />
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
                    <WorkspacesPreview
                        // previewDisabled={previewDisabled}
                        previewDisabled={true}
                        // preview={() => previewBatchSpec(debouncedCode)}
                        preview={() => alert('preview')}
                        batchSpecStale={true}
                        // batchSpecStale={
                        //     isBatchSpecStale || isWorkspacesPreviewInProgress || resolutionState === 'CANCELED'
                        // }
                        hasPreviewed={false}
                        excludeRepo={editor.excludeRepo}
                        cancel={() => alert('cancel')}
                        isWorkspacesPreviewInProgress={false}
                        resolutionState="CANCELED"
                        workspacesConnection={{} as any}
                        importingChangesetsConnection={{} as any}
                        setFilters={() => alert('setFilters')}
                        // hasPreviewed={hasPreviewed}
                        // excludeRepo={excludeRepo}
                        // cancel={cancel}
                        // isWorkspacesPreviewInProgress={isWorkspacesPreviewInProgress}
                        // resolutionState={resolutionState}
                        // workspacesConnection={workspacesConnection}
                        // importingChangesetsConnection={importingChangesetsConnection}
                        // setFilters={setFilters}
                    />
                </div>
            </Panel>
        </div>
    )
}
