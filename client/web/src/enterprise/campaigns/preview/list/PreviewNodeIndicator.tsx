import React from 'react'
import { ChangesetApplyPreviewFields } from '../../../../graphql-operations'

const containerClassName = 'd-none d-sm-flex flex-column align-items-center justify-content-center align-self-stretch'

export interface PreviewNodeIndicatorProps {
    node: ChangesetApplyPreviewFields
}

export const PreviewNodeIndicator: React.FunctionComponent<PreviewNodeIndicatorProps> = ({ node }) => {
    switch (node.targets.__typename) {
        case 'HiddenApplyPreviewTargetsAttach':
        case 'VisibleApplyPreviewTargetsAttach':
            return (
                <div className={containerClassName}>
                    <span className="preview-node-indicator__attach-dot">&nbsp;</span>
                    <span className="preview-node-indicator__attach-icon">+</span>
                    <span className="preview-node-indicator__attach-dot">&nbsp;</span>
                </div>
            )
        case 'HiddenApplyPreviewTargetsUpdate':
        case 'VisibleApplyPreviewTargetsUpdate':
            if (node.__typename === 'HiddenChangesetApplyPreview' || node.operations.length === 0) {
                // If no operations, no update :P
                return <div />
            }
            return (
                <div className={containerClassName}>
                    <span className="preview-node-indicator__update-dot">&nbsp;</span>
                    <span className="preview-node-indicator__update-icon">&bull;</span>
                    <span className="preview-node-indicator__update-dot">&nbsp;</span>
                </div>
            )
        case 'HiddenApplyPreviewTargetsDetach':
        case 'VisibleApplyPreviewTargetsDetach':
            return (
                <div className={containerClassName}>
                    <span className="preview-node-indicator__detach-dot">&nbsp;</span>
                    <span className="preview-node-indicator__detach-icon">-</span>
                    <span className="preview-node-indicator__detach-dot">&nbsp;</span>
                </div>
            )
    }
}
