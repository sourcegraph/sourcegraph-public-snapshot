import React from 'react'
import { ChangesetSpecType, HiddenChangesetApplyPreviewFields } from '../../../graphql-operations'
import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import { ChangesetSpecAction } from './ChangesetSpecAction'

export interface HiddenChangesetSpecNodeProps {
    node: HiddenChangesetApplyPreviewFields
}

export const HiddenChangesetSpecNode: React.FunctionComponent<HiddenChangesetSpecNodeProps> = ({ node }) => (
    <>
        <span className="d-none d-sm-block" />
        <ChangesetSpecAction node={node} className="hidden-changeset-spec-node__action" />
        <div className="d-flex flex-column hidden-changeset-spec-node__information">
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
