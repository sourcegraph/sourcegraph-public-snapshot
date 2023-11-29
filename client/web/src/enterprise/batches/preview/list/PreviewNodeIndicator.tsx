import React from 'react'

import classNames from 'classnames'

import { type ChangesetApplyPreviewFields, ChangesetSpecOperation } from '../../../../graphql-operations'
import { ChangesetAddedIcon, ChangesetModifiedIcon, ChangesetRemovedIcon } from '../icons'

import styles from './PreviewNodeIndicator.module.scss'

const containerClassName = classNames(
    styles.previewNodeIndicatorContainer,
    'd-none d-sm-flex flex-column align-items-center justify-content-center align-self-stretch'
)

export interface PreviewNodeIndicatorProps {
    node: ChangesetApplyPreviewFields
}

export const PreviewNodeIndicator: React.FunctionComponent<React.PropsWithChildren<PreviewNodeIndicatorProps>> = ({
    node,
}) => {
    switch (node.targets.__typename) {
        case 'HiddenApplyPreviewTargetsAttach':
        case 'VisibleApplyPreviewTargetsAttach': {
            return (
                <div className={containerClassName}>
                    <span className={styles.previewNodeIndicatorAttachBar}>&nbsp;</span>
                    <span
                        className={classNames(
                            styles.previewNodeIndicatorAttachIcon,
                            'd-flex justify-content-center align-items-center'
                        )}
                    >
                        <ChangesetAddedIcon />
                    </span>
                    <span className={styles.previewNodeIndicatorAttachBar}>&nbsp;</span>
                </div>
            )
        }
        case 'HiddenApplyPreviewTargetsUpdate':
        case 'VisibleApplyPreviewTargetsUpdate': {
            if (node.__typename === 'HiddenChangesetApplyPreview' || node.operations.length === 0) {
                // If no operations, no update :P
                return <div />
            }
            if (node.operations.includes(ChangesetSpecOperation.REATTACH)) {
                return (
                    <div className={containerClassName}>
                        <span className={styles.previewNodeIndicatorAttachBar}>&nbsp;</span>
                        <span
                            className={classNames(
                                styles.previewNodeIndicatorAttachIcon,
                                'd-flex justify-content-center align-items-center'
                            )}
                        >
                            <ChangesetAddedIcon />
                        </span>
                        <span className={styles.previewNodeIndicatorAttachBar}>&nbsp;</span>
                    </div>
                )
            }
            return (
                <div className={containerClassName}>
                    <span className={styles.previewNodeIndicatorUpdateBar}>&nbsp;</span>
                    <span
                        className={classNames(
                            styles.previewNodeIndicatorUpdateIcon,
                            'd-flex justify-content-center align-items-center'
                        )}
                    >
                        <ChangesetModifiedIcon />
                    </span>
                    <span className={styles.previewNodeIndicatorUpdateBar}>&nbsp;</span>
                </div>
            )
        }
        case 'HiddenApplyPreviewTargetsDetach':
        case 'VisibleApplyPreviewTargetsDetach': {
            return (
                <div className={containerClassName}>
                    <span className={styles.previewNodeIndicatorDetachBar}>&nbsp;</span>
                    <span
                        className={classNames(
                            styles.previewNodeIndicatorDetachIcon,
                            'd-flex justify-content-center align-items-center'
                        )}
                    >
                        <ChangesetRemovedIcon />
                    </span>
                    <span className={styles.previewNodeIndicatorDetachBar}>&nbsp;</span>
                </div>
            )
        }
    }
}
