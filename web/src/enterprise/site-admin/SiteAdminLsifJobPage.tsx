import * as GQL from '../../../../shared/src/graphql/schema'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import ClockOutlineIcon from 'mdi-react/ClockOutlineIcon'
import React, { FunctionComponent, useEffect, useMemo } from 'react'
import { asError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { catchError } from 'rxjs/operators'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchLsifJob } from './backend'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { PageTitle } from '../../components/PageTitle'
import { sortBy } from 'lodash'
import { RouteComponentProps } from 'react-router'
import { Timestamp } from '../../components/time/Timestamp'
import { useObservable } from '../../util/useObservable'
import { ErrorAlert } from '../../components/alerts'

const JobArguments: FunctionComponent<{ args: { [name: string]: string } }> = ({ args }) => (
    <table className="job-arguments w-100 mb-0 table table-sm">
        <tbody>
            {sortBy(Object.entries(args), ([key]) => key).map(([key, value]) => (
                <tr key={key} className="job-arguments__row">
                    <td className="job-arguments__cell">{key}</td>
                    <td className="job-arguments__cell">{value}</td>
                </tr>
            ))}
        </tbody>
    </table>
)

interface Props extends RouteComponentProps<{ id: string }> {}

/**
 * A page displaying metadata about an LSIF job.
 */
export const SiteAdminLsifJobPage: FunctionComponent<Props> = ({
    match: {
        params: { id },
    },
}) => {
    useEffect(() => eventLogger.logViewEvent('SiteAdminLsifJob'))

    const jobOrError = useObservable(
        useMemo(() => fetchLsifJob({ id }).pipe(catchError((error): [ErrorLike] => [asError(error)])), [id])
    )

    return (
        <div className="site-admin-lsif-job-page w-100">
            <PageTitle title="LSIF jobs - Admin" />
            {!jobOrError ? (
                <LoadingSpinner className="icon-inline" />
            ) : isErrorLike(jobOrError) ? (
                <div className="alert alert-danger">
                    <ErrorAlert prefix="Error loading LSIF job" error={jobOrError} />
                </div>
            ) : (
                <>
                    <div className="mt-3 mb-1">
                        <h2 className="mb-0">{lsifJobDescription(jobOrError)}</h2>
                    </div>

                    {jobOrError.state === GQL.LSIFJobState.PROCESSING ? (
                        <div className="alert alert-primary mb-4 mt-3">
                            <LoadingSpinner className="icon-inline" /> Job is currently being processed...
                        </div>
                    ) : jobOrError.state === GQL.LSIFJobState.COMPLETED ? (
                        <div className="alert alert-success mb-4 mt-3">
                            <CheckIcon className="icon-inline" /> Job completed successfully.
                        </div>
                    ) : jobOrError.state === GQL.LSIFJobState.ERRORED ? (
                        <div className="alert alert-danger mb-4 mt-3">
                            <AlertCircleIcon className="icon-inline" /> Job failed to complete:{' '}
                            <code>{jobOrError.failure && jobOrError.failure.summary}</code>
                        </div>
                    ) : (
                        <div className="alert alert-primary mb-4 mt-3">
                            <ClockOutlineIcon className="icon-inline" /> Job is queued.
                        </div>
                    )}

                    <table className="table">
                        <tbody>
                            <tr>
                                <td>Queued</td>
                                <td>
                                    <Timestamp date={jobOrError.queuedAt} noAbout={true} />
                                </td>
                            </tr>

                            <tr>
                                <td>Began processing</td>
                                <td>
                                    {jobOrError.startedAt ? (
                                        <Timestamp date={jobOrError.startedAt} noAbout={true} />
                                    ) : (
                                        <span className="text-muted">Job has not yet started.</span>
                                    )}
                                </td>
                            </tr>

                            <tr>
                                <td>
                                    {jobOrError.state === GQL.LSIFJobState.ERRORED && jobOrError.completedOrErroredAt
                                        ? 'Failed'
                                        : 'Finished'}{' '}
                                    processing
                                </td>
                                <td>
                                    {jobOrError.completedOrErroredAt ? (
                                        <Timestamp date={jobOrError.completedOrErroredAt} noAbout={true} />
                                    ) : (
                                        <span className="text-muted">Job has not yet completed.</span>
                                    )}
                                </td>
                            </tr>

                            {Object.keys(jobOrError.arguments || {}).length > 0 && (
                                <tr>
                                    <td>Arguments</td>
                                    <td className="pt-0 pb-0 pr-0">
                                        <JobArguments args={jobOrError.arguments} />
                                    </td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </>
            )}
        </div>
    )
}

/**
 * Construct a meaningful description from the job name and args.
 *
 * @param job The job instance.
 */
function lsifJobDescription(job: GQL.ILSIFJob): string {
    if (job.type === 'convert') {
        const {
            repository,
            commit,
            root,
        }: {
            repository: string
            commit: string
            root: string
        } = job.arguments

        return `Convert upload for ${repository} at ${commit.substring(0, 7)}${root === '' ? '' : `, ${root}`}`
    }

    const internalJobs: { [name: string]: string } = {
        'clean-old-jobs': 'Purge old job data from LSIF work queue',
        'clean-failed-jobs': 'Clean old failed job uploads from disk',
        'update-tips': 'Refresh current uploads',
    }

    if (internalJobs[job.type]) {
        return `Internal job: ${internalJobs[job.type]}`
    }

    return `Unknown job type ${job.type}`
}
