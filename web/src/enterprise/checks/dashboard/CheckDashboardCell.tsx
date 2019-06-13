import classNames from 'classnames'
import AlertIcon from 'mdi-react/AlertIcon'
import React, { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../shared/src/util/strings'
import { parseJSON } from '../../../settings/configuration'
import { Changeset, computeChangesets, getChangesetExternalStatus } from '../../threads/detail/backend'

interface Props extends ExtensionsControllerProps {
    thread: GQL.IDiscussionThread
    repo: string

    paddingClassName?: string
}

const LOADING: 'loading' = 'loading'

/**
 * A cell in the checks dashboard.
 */
export const CheckDashboardCell: React.FunctionComponent<Props> = ({
    thread,
    repo,
    paddingClassName = '',
    extensionsController,
}) => {
    const threadSettings = useMemo(() => parseJSON(thread.settings), [thread])

    const [changesetsOrError, setChangesetsOrError] = useState<typeof LOADING | Changeset[] | ErrorLike>(LOADING)
    // tslint:disable-next-line: no-floating-promises
    useEffect(() => {
        const subscriptions = new Subscription()
        subscriptions.add(
            computeChangesets(extensionsController, thread, threadSettings, { repo })
                .pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
                .subscribe(setChangesetsOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [thread.id, threadSettings, repo, extensionsController, thread])

    return changesetsOrError === LOADING ? (
        <span />
    ) : isErrorLike(changesetsOrError) ? (
        <div className={paddingClassName}>
            <AlertIcon className="icon-inline p-2 text-danger" data-tooltip={changesetsOrError.message} />
        </div>
    ) : (
        <div className={`d-flex w-100 h-100 align-items-center p-0`}>
            {changesetsOrError.length === 0 ? (
                <Link
                    to={`${thread.url}/inbox?q=${repo}`}
                    className={`d-block text-decoration-none bg-success flex-1 ${paddingClassName}`}
                >
                    &nbsp;
                </Link>
            ) : (
                changesetsOrError.map((changeset, i) => {
                    const { status } = getChangesetExternalStatus(changeset)
                    return (
                        <Link
                            key={i}
                            to={`${thread.url}/inbox?q=${repo}`}
                            className={classNames('d-block', 'text-decoration-none', 'flex-1', paddingClassName, {
                                'bg-success': changeset.fileDiffs.length === 0 && status === 'merged',
                                'bg-warning': changeset.fileDiffs.length > 0 && status === 'merged',
                                'bg-danger': changeset.fileDiffs.length > 0 && status !== 'merged',
                            })}
                            data-tooltip={`${status === 'merged' ? '1 PR merged so far, ' : ''}${
                                changeset.fileDiffs.length
                            } ${pluralize('outstanding problem', changeset.fileDiffs.length)}`}
                        >
                            &nbsp;
                        </Link>
                    )
                })
            )}
        </div>
    )
}
