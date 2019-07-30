import { NotificationType } from '@sourcegraph/extension-api-classes'
import { Diagnostic } from '@sourcegraph/extension-api-types'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { flatten, uniq } from 'lodash'
import AlertIcon from 'mdi-react/AlertIcon'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { Redirect } from 'react-router'
import { UncontrolledPopover } from 'reactstrap'
import { filter, first, map } from 'rxjs/operators'
import { fromDiagnostic } from '../../../../../../../shared/src/api/types/diagnostic'
import { RepositoryIcon } from '../../../../../../../shared/src/components/icons'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../../../shared/src/util/strings'
import { parseRepoURI } from '../../../../../../../shared/src/util/url'
import { DiffStat } from '../../../../../repo/compare/DiffStat'
import { DiffIcon, ZapIcon } from '../../../../../util/octicons'
import { createCampaignFromDiffs } from '../../../../changesetsOLD/preview/backend'
import { ChangesetTargetButtonDropdown } from '../../../../tasks/list/item/ChangesetTargetButtonDropdown'
import { diagnosticQueryMatcher, getDiagnosticInfos } from '../../../../threadsOLD/detail/backend'
import { computeDiff, computeDiffStat, FileDiff } from '../../../../threadsOLD/detail/changes/computeDiff'
import { ChangesetPlanProps } from '../useChangesetPlan'

interface Props extends Pick<ChangesetPlanProps, 'changesetPlan'>, ExtensionsControllerProps {
    className?: string
}

const LOADING = 'loading' as const

const DEBUG = localStorage.getItem('debug') !== null

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

    const isEmpty = changesetPlan.operations.length === 0

    const [fileDiffsOrError, setFileDiffsOrError] = useState<typeof LOADING | FileDiff[] | ErrorLike>(LOADING)
    useEffect(() => {
        const do2 = async () => {
            setFileDiffsOrError(LOADING)
            try {
                setFileDiffsOrError(
                    await computeDiff(
                        extensionsController,
                        flatten(
                            await Promise.all(
                                changesetPlan.operations.map(async op => {
                                    const diagnostics: (Diagnostic | null)[] = op.diagnostics
                                        ? await getDiagnosticInfos(extensionsController, op.diagnostics.type)
                                              .pipe(
                                                  first() /*TODO!(sqs) remove first, make reactive*/,
                                                  map(diagnostics =>
                                                      op.diagnostics
                                                          ? diagnostics.filter(diagnosticQueryMatcher(op.diagnostics))
                                                          : diagnostics
                                                  ),
                                                  map(diagnostics => diagnostics.map(fromDiagnostic))
                                              )
                                              .toPromise()
                                        : [null]
                                    return diagnostics.map(diagnostic => ({
                                        actionEditCommand: op.editCommand,
                                        diagnostic,
                                    }))
                                })
                            )
                        )
                    )
                )
            } catch (err) {
                setFileDiffsOrError(asError(err))
            }
        }
        do2()
    }, [changesetPlan.operations, extensionsController])
    const diffStat =
        fileDiffsOrError !== LOADING && !isErrorLike(fileDiffsOrError) ? computeDiffStat(fileDiffsOrError) : null

    const [createdThreadOrLoading, setCreatedThreadOrLoading] = useState<
        null | typeof LOADING | Pick<GQL.ICampaign, 'id' | 'url'>
    >(null)
    const onCreateThreadClick = useCallback(() => {
        const do2 = async () => {
            setCreatedThreadOrLoading(LOADING)
            try {
                if (fileDiffsOrError === LOADING || isErrorLike(fileDiffsOrError)) {
                    throw new Error('file diffs not available')
                }
                setCreatedThreadOrLoading(
                    await createCampaignFromDiffs(fileDiffsOrError, {
                        name: `Fix eslint issues`, // TODO!(sqs): un-hardcode eslint
                        description: `Fixes eslint issues in ${fileDiffsOrError.length} ${pluralize(
                            'file',
                            fileDiffsOrError.length
                        )}`,
                        preview: true,
                        rules: JSON.stringify(changesetPlan.operations),
                    })
                )
            } catch (err) {
                setCreatedThreadOrLoading(null)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error creating changeset: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        }
        do2()
    }, [changesetPlan, extensionsController, fileDiffsOrError])

    const repositoriesAffected =
        fileDiffsOrError !== LOADING &&
        !isErrorLike(fileDiffsOrError) &&
        uniq(fileDiffsOrError.map(fileDiff => parseRepoURI(fileDiff.newPath || fileDiff.oldPath!).repoName)).length

    return (
        <div className={`diagnostics-changesets-bar ${flashBorderClassName} ${flashBackgroundClassName} ${className}`}>
            <div className="container py-4 d-flex align-items-center">
                {createdThreadOrLoading === null || createdThreadOrLoading === LOADING ? (
                    <ChangesetTargetButtonDropdown
                        onClick={onCreateThreadClick}
                        buttonClassName="btn-success"
                        className="mr-4"
                        loading={createdThreadOrLoading === LOADING}
                        disabled={
                            isEmpty ||
                            fileDiffsOrError === LOADING ||
                            isErrorLike(fileDiffsOrError) ||
                            createdThreadOrLoading === LOADING
                        }
                    />
                ) : (
                    <Redirect to={createdThreadOrLoading.url} push={true} />
                )}

                {!isEmpty ? (
                    <div className={`flex-1 d-flex align-items-center`}>
                        <span className="mr-4">
                            <ZapIcon className="icon-inline" /> <strong>{changesetPlan.operations.length}</strong>{' '}
                            <span className="text-muted">
                                {pluralize('operation', changesetPlan.operations.length)}
                            </span>
                        </span>
                        <div className="mr-4">
                            <DiffIcon className="icon-inline" />{' '}
                            {fileDiffsOrError === LOADING ? (
                                <LoadingSpinner className="icon-inline" />
                            ) : isErrorLike(fileDiffsOrError) ? (
                                <span title={fileDiffsOrError.message}>
                                    <AlertIcon className="icon-inline text-danger" />
                                </span>
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
                        {typeof repositoriesAffected === 'number' && (
                            <span className="mr-4">
                                <RepositoryIcon className="icon-inline" /> {/* TODO!(sqs): fake computation */}
                                <strong>{repositoriesAffected}</strong>{' '}
                                <span className="text-muted">
                                    {pluralize('repository affected', repositoriesAffected, 'repositories affected')}
                                </span>
                            </span>
                        )}
                        <div className="flex-1" />
                    </div>
                ) : (
                    <div className={`text-muted`}>Select actions to include in changeset...</div>
                )}
                <div className="flex-1" />
                {DEBUG && (
                    <>
                        <button id="diagnostic-changesets-bar__popover-trigger" type="button" className="btn btn-link">
                            <InformationOutlineIcon className="icon-inline" />
                        </button>
                        <UncontrolledPopover
                            trigger="click"
                            placement="top"
                            target="diagnostic-changesets-bar__popover-trigger"
                        >
                            <div className="card bg-body">
                                <h5 className="card-header">
                                    Changeset <span className="badge badge-secondary">Debug</span>
                                </h5>
                                <pre
                                    style={{
                                        overflow: 'auto',
                                        width: '40vw',
                                        height: '50vh',
                                        fontSize: '11px',
                                    }}
                                    className="card-body"
                                >
                                    {JSON.stringify(changesetPlan, null, 2)}
                                </pre>
                            </div>
                        </UncontrolledPopover>
                    </>
                )}
            </div>
        </div>
    )
}
