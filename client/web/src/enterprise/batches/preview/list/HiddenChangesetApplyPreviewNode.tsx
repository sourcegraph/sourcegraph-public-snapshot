import classNames from 'classnames'
import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import React from 'react'

import { ChangesetState } from '@sourcegraph/shared/src/graphql-operations'

import { ChangesetSpecType, HiddenChangesetApplyPreviewFields } from '../../../../graphql-operations'
import { ChangesetStatusCell } from '../../detail/changesets/ChangesetStatusCell'

import styles from './HiddenChangesetApplyPreviewNode.module.scss'
import { PreviewActions } from './PreviewActions'
import { PreviewNodeIndicator } from './PreviewNodeIndicator'

export interface HiddenChangesetApplyPreviewNodeProps {
    node: HiddenChangesetApplyPreviewFields
}

export const HiddenChangesetApplyPreviewNode: React.FunctionComponent<HiddenChangesetApplyPreviewNodeProps> = ({
    node,
}) => (
    <>
        <span className={classNames(styles.hiddenChangesetApplyPreviewNodeListCell, 'd-none d-sm-block')} />
        <HiddenChangesetApplyPreviewNodeStatusCell
            node={node}
            className={classNames(
                styles.hiddenChangesetApplyPreviewNodeListCell,
                styles.hiddenChangesetApplyPreviewNodeCurrentState,
                'd-block d-sm-flex'
            )}
        />
        <PreviewNodeIndicator node={node} />
        <PreviewActions
            node={node}
            className={classNames(
                styles.hiddenChangesetApplyPreviewNodeListCell,
                styles.hiddenChangesetApplyPreviewNodeAction
            )}
        />
        <div
            className={classNames(
                styles.hiddenChangesetApplyPreviewNodeListCell,
                styles.hiddenChangesetApplyPreviewNodeInformation,
                ' d-flex flex-column'
            )}
        >
            <h3 className="text-muted">
                {node.targets.__typename === 'HiddenApplyPreviewTargetsAttach' ||
                node.targets.__typename === 'HiddenApplyPreviewTargetsUpdate' ? (
                    <>
                        {node.targets.changesetSpec.type === ChangesetSpecType.EXISTING && (
                            <>Import changeset from a private repository</>
                        )}
                        {node.targets.changesetSpec.type === ChangesetSpecType.BRANCH && (
                            <>Create changeset in a private repository</>
                        )}
                    </>
                ) : (
                    <>Detach changeset in a private repository</>
                )}
            </h3>
            <span className="text-danger">
                No action will be taken on apply.{' '}
                <InfoCircleOutlineIcon
                    className="icon-inline"
                    data-tooltip="You have no permissions to access this repository."
                />
            </span>
        </div>
        <span />
        <span />
    </>
)

const HiddenChangesetApplyPreviewNodeStatusCell: React.FunctionComponent<
    HiddenChangesetApplyPreviewNodeProps & { className?: string }
> = ({ node, className }) => {
    if (node.targets.__typename === 'HiddenApplyPreviewTargetsAttach') {
        return <ChangesetStatusCell state={ChangesetState.UNPUBLISHED} className={className} />
    }
    return <ChangesetStatusCell state={node.targets.changeset.state} className={className} />
}
