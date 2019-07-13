import { NotificationType } from '@sourcegraph/extension-api-classes'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AlertIcon from 'mdi-react/AlertIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { DropdownToggle } from 'reactstrap'
import { toAction } from '../../../../../../../shared/src/api/types/action'
import { RepositoryIcon } from '../../../../../../../shared/src/components/icons'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../../../shared/src/util/strings'
import { DiffStat } from '../../../../../repo/compare/DiffStat'
import { DiffIcon, ZapIcon } from '../../../../../util/octicons'
import { useEffectAsync } from '../../../../../util/useEffectAsync'
import { ChangesetIcon } from '../../../../changesets/icons'
import { createChangesetFromCodeAction } from '../../../../changesets/preview/backend'
import { ChangesetButtonOrLinkExistingChangeset } from '../../../../tasks/list/item/ChangesetButtonOrLink'
import { ChangesetTargetButtonDropdown } from '../../../../tasks/list/item/ChangesetTargetButtonDropdown'
import { computeDiff, computeDiffStat, FileDiff } from '../../../../threads/detail/changes/computeDiff'
import { ChangesetPlanProps } from '../useChangesetPlan'
import { DiagnosticsBatchActions } from './DiagnosticsBatchActions'

interface Props extends ChangesetPlanProps, ExtensionsControllerProps {
    className?: string
}

const LOADING = 'loading' as const

/**
 * A bar displaying the changesets related to a set of diagnostics, plus a preview of and statistics
 * about a new changeset that is being created.
 */
export const DiagnosticsChangesetsBar: React.FunctionComponent<Props> = ({
    changesetPlan,
    extensionsController,
    className = '',
}) => {
    const [justChanged, setJustChanged] = useState(false)
    const [lastChangesetPlan, setLastChangesetPlan] = useState(changesetPlan)
    if (changesetPlan !== lastChangesetPlan) {
        setJustChanged(true)
        setLastChangesetPlan(changesetPlan)
        setTimeout(() => setJustChanged(false), 500)
    }
    const flashBorderClassName = justChanged ? 'diagnostics-changesets-bar--flash-border' : ''
    const flashBackgroundClassName = justChanged ? 'diagnostics-changesets-bar--flash-bg' : ''

    const [createdThreadOrLoading, setCreatedThreadOrLoading] = useState<ChangesetButtonOrLinkExistingChangeset>(
        LOADING
    )
    const onCreateThreadClick = useCallback(async () => {
        setCreatedThreadOrLoading(PENDING_CREATION)
        try {
            const action = selectedAction
            if (!action) {
                throw new Error('no active code action')
            }
            setCreatedThreadOrLoading(
                await createChangesetFromCodeAction({ extensionsController }, null, action, {
                    status: GQL.ThreadStatus.PREVIEW,
                })
            )
        } catch (err) {
            setCreatedThreadOrLoading(null)
            extensionsController.services.notifications.showMessages.next({
                message: `Error creating changeset: ${err.message}`,
                type: NotificationType.Error,
            })
        }
    }, [extensionsController])

    const isEmpty = changesetPlan.operations.length === 0 || changesetPlan.operations[0].diagnosticActions.length === 0

    const [fileDiffsOrError, setFileDiffsOrError] = useState<typeof LOADING | FileDiff[] | ErrorLike>(LOADING)
    useEffectAsync(async () => {
        setFileDiffsOrError(LOADING)
        try {
            setFileDiffsOrError(
                await computeDiff(
                    extensionsController,
                    changesetPlan.operations.flatMap(op => op.diagnosticActions.map(({ action }) => toAction(action)))
                )
            )
        } catch (err) {
            setFileDiffsOrError(asError(err))
        }
    }, [changesetPlan.operations, extensionsController])
    const diffStat =
        fileDiffsOrError !== LOADING && !isErrorLike(fileDiffsOrError) ? computeDiffStat(fileDiffsOrError) : null

    return (
        <div className={`diagnostics-changesets-bar ${flashBorderClassName} ${flashBackgroundClassName} ${className}`}>
            <div className="container py-4 d-flex align-items-center">
                <ChangesetTargetButtonDropdown
                    onClick={() => {
                        throw new Error('TODO!(sqs)')
                    }}
                    showAddToExistingChangeset={true}
                    buttonClassName="btn-success"
                    className="mr-4"
                    disabled={isEmpty}
                />

                {!isEmpty ? (
                    <div className={`flex-1 d-flex align-items-center`}>
                        <span className="mr-4">
                            <ZapIcon className="icon-inline" />{' '}
                            <strong>{changesetPlan.operations[0].diagnosticActions.length}</strong>{' '}
                            <span className="text-muted">
                                {pluralize('action', changesetPlan.operations[0].diagnosticActions.length)}
                            </span>
                        </span>
                        <div className="mr-4">
                            <DiffIcon className="icon-inline" />{' '}
                            {fileDiffsOrError === LOADING ? (
                                <LoadingSpinner className="icon-inline" />
                            ) : isErrorLike(fileDiffsOrError) ? (
                                <AlertIcon className="icon-inline text-danger" title={fileDiffsOrError.message} />
                            ) : (
                                <>
                                    <strong>{fileDiffsOrError.length}</strong>{' '}
                                    <span className="text-muted">
                                        {pluralize('file changed', fileDiffsOrError.length, 'files changed')}
                                    </span>
                                    {diffStat && <DiffStat {...diffStat} className="ml-3 d-inline-flex" />}
                                </>
                            )}
                        </div>
                        <span className="mr-4">
                            <RepositoryIcon className="icon-inline" /> {/* TODO!(sqs): fake computation */}
                            <strong>
                                {1 + Math.floor(changesetPlan.operations[0].diagnosticActions.length / 5)}
                            </strong>{' '}
                            <span className="text-muted">
                                {pluralize(
                                    'repository affected',
                                    1 + Math.floor(changesetPlan.operations[0].diagnosticActions.length / 5),
                                    'repositories affected'
                                )}
                            </span>
                        </span>
                        <div className="flex-1"></div>
                    </div>
                ) : (
                    <div className={`text-muted`}>Select actions to include in changeset...</div>
                )}
            </div>
        </div>
    )
}
