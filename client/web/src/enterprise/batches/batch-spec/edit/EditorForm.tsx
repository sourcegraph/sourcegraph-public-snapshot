import React, { useContext } from 'react'

import { BatchSpecContext } from '../BatchSpecContext'

import { LibraryPane } from './library/LibraryPane'

import styles from './EditorForm.module.scss'

export const EditorForm: React.FunctionComponent<React.PropsWithChildren<{}>> = () => {
    const { batchChange, batchSpec, editor } = useContext(BatchSpecContext)
    return (
        <div className={styles.form}>
            <LibraryPane name={batchChange.name} onReplaceItem={() => alert('hi')} />
            <div className={styles.editorContainer}>
                <h4 className={styles.header}>Batch spec</h4>

                {/* <MonacoBatchSpecEditor
                        batchChangeName={batchChange.name}
                        className={styles.editor}
                        isLightTheme={isLightTheme}
                        value={code}
                        onChange={clearErrorsAndHandleCodeChange}
                    />
                    <EditorFeedbackPanel
                        errors={{
                            codeUpdate: codeErrors.update,
                            codeValidation: codeErrors.validation,
                            preview: previewError,
                            execute: executeError,
                        }}
                    />

                    {isDownloadSpecModalOpen && !downloadSpecModalDismissed ? (
                        <DownloadSpecModal
                            name={batchChange.name}
                            originalInput={code}
                            isLightTheme={isLightTheme}
                            setDownloadSpecModalDismissed={setDownloadSpecModalDismissed}
                            setIsDownloadSpecModalOpen={setIsDownloadSpecModalOpen}
                        />
                    ) : null} */}
            </div>
            {/* <Panel
                    className="d-flex"
                    defaultSize={500}
                    minSize={405}
                    maxSize={1400}
                    position="right"
                    storageKey={WORKSPACES_PREVIEW_SIZE}
                >
                    <div className={styles.workspacesPreviewContainer}>
                        <WorkspacesPreview
                            previewDisabled={previewDisabled}
                            preview={() => previewBatchSpec(debouncedCode)}
                            batchSpecStale={
                                isBatchSpecStale || isWorkspacesPreviewInProgress || resolutionState === 'CANCELED'
                            }
                            hasPreviewed={hasPreviewed}
                            excludeRepo={excludeRepo}
                            cancel={cancel}
                            isWorkspacesPreviewInProgress={isWorkspacesPreviewInProgress}
                            resolutionState={resolutionState}
                            workspacesConnection={workspacesConnection}
                            importingChangesetsConnection={importingChangesetsConnection}
                            setFilters={setFilters}
                        />
                    </div>
                </Panel> */}
        </div>
    )
}
