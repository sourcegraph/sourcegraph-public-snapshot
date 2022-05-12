import React from 'react'

import classNames from 'classnames'
import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'

import { ChangesetState } from '@sourcegraph/shared/src/graphql-operations'
import { Icon, Typography } from '@sourcegraph/wildcard'

import { InputTooltip } from '../../../../components/InputTooltip'
import { ChangesetSpecType, HiddenChangesetApplyPreviewFields } from '../../../../graphql-operations'
import { ChangesetStatusCell } from '../../detail/changesets/ChangesetStatusCell'

import { PreviewActions } from './PreviewActions'
import { PreviewNodeIndicator } from './PreviewNodeIndicator'

import styles from './HiddenChangesetApplyPreviewNode.module.scss'

export interface HiddenChangesetApplyPreviewNodeProps {
    node: HiddenChangesetApplyPreviewFields
}

export const HiddenChangesetApplyPreviewNode: React.FunctionComponent<
    React.PropsWithChildren<HiddenChangesetApplyPreviewNodeProps>
> = ({ node }) => (
    <>
        <span className={classNames(styles.hiddenChangesetApplyPreviewNodeListCell, 'd-none d-sm-block')} />
        <div className="p-2">
            <InputTooltip
                id="select-changeset-hidden"
                type="checkbox"
                checked={false}
                disabled={true}
                tooltip="You do not have permission to publish to this repository."
            />
        </div>
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
            <Typography.H3 className="text-muted">
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
            </Typography.H3>
            <span className="text-danger">
                No action will be taken on apply.{' '}
                <Icon data-tooltip="You have no permissions to access this repository." as={InfoCircleOutlineIcon} />
            </span>
        </div>
        <span />
        <span />
    </>
)

const HiddenChangesetApplyPreviewNodeStatusCell: React.FunctionComponent<
    React.PropsWithChildren<HiddenChangesetApplyPreviewNodeProps & { className?: string }>
> = ({ node, className }) => {
    if (node.targets.__typename === 'HiddenApplyPreviewTargetsAttach') {
        return <ChangesetStatusCell state={ChangesetState.UNPUBLISHED} className={className} />
    }
    return <ChangesetStatusCell state={node.targets.changeset.state} className={className} />
}
