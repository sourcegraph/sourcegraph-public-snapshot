import React from 'react'
import { ChangesetSpecType, HiddenChangesetApplyPreviewFields } from '../../../../graphql-operations'
import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import { PreviewAction } from './PreviewAction'

export interface HiddenChangesetApplyPreviewNodeProps {
    node: HiddenChangesetApplyPreviewFields
}

export const HiddenChangesetApplyPreviewNode: React.FunctionComponent<HiddenChangesetApplyPreviewNodeProps> = ({
    node,
}) => (
    <>
        <span className="d-none d-sm-block" />
        <PreviewAction node={node} className="hidden-changeset-apply-preview-node__action" />
        <div className="d-flex flex-column hidden-changeset-apply-preview-node__information">
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
    </>
)
