import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import UnfoldLessVerticalIcon from 'mdi-react/UnfoldLessVerticalIcon'
import UnfoldMoreVerticalIcon from 'mdi-react/UnfoldMoreVerticalIcon'
import React, { useCallback, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { RepoLink } from '../../../../../shared/src/components/RepoLink'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { fetchDiscussionThreads } from '../../../discussions/backend'
import { useEffectAsync } from '../../../util/useEffectAsync'
import { ChecksAreaTitle } from '../components/ChecksAreaTitle'
import { CheckDashboardCell } from './CheckDashboardCell'

interface Props extends ExtensionsControllerProps {
    location: H.Location
}

const REPOS = [
    'github.com/sourcegraph/sourcegraph',
    'github.com/sourcegraph/about',
    'github.com/sourcegraph/codeintellify',
    'github.com/sourcegraph/react-loading-spinner',
    'github.com/sourcegraph/sourcegraph',
    'github.com/lyft/pipelines',
    'github.com/lyft/amundsenfrontendlibrary',
]

const LOADING: 'loading' = 'loading'

const LOCAL_STORAGE_KEY = 'ChecksDashboardPage-container'

/**
 * A dashboard for checks.
 */
export const ChecksDashboardPage: React.FunctionComponent<Props> = ({ location, ...props }) => {
    const initialIsExpanded = useMemo(() => localStorage.getItem(LOCAL_STORAGE_KEY) !== null, [])
    const [isExpanded, setIsExpanded] = useState(initialIsExpanded)
    const toggleIsExpanded = useCallback(() => {
        setIsExpanded(!isExpanded)
        if (isExpanded) {
            localStorage.removeItem(LOCAL_STORAGE_KEY)
        } else {
            localStorage.setItem(LOCAL_STORAGE_KEY, 'expanded')
        }
    }, [isExpanded])

    const [threadsOrError, setThreadsOrError] = useState<typeof LOADING | GQL.IDiscussionThreadConnection | ErrorLike>(
        LOADING
    )
    useEffectAsync(async () => {
        try {
            setThreadsOrError(await fetchDiscussionThreads({}).toPromise())
        } catch (err) {
            setThreadsOrError(asError(err))
        }
    }, [])

    return (
        <div className={`${isExpanded ? 'container-fluid' : 'container'} mt-3`}>
            <ChecksAreaTitle>
                <button
                    type="button"
                    className="btn btn-link text-decoration-none"
                    data-tooltip={isExpanded ? 'Exit widescreen view' : 'Enter widescreen view'}
                    onClick={toggleIsExpanded}
                >
                    {isExpanded ? (
                        <UnfoldLessVerticalIcon className="icon-inline" />
                    ) : (
                        <UnfoldMoreVerticalIcon className="icon-inline" />
                    )}
                </button>
            </ChecksAreaTitle>
            {threadsOrError === LOADING ? (
                <LoadingSpinner className="mt-3 mx-auto" />
            ) : isErrorLike(threadsOrError) ? (
                <div className="alert alert-danger">{threadsOrError.message}</div>
            ) : (
                <table className="table table-bordered border-0 table-responsive-md">
                    <thead>
                        <tr>
                            <th className="border-top-0 border-left-0" />
                            {threadsOrError.nodes.map((check, i) => (
                                <th key={i}>
                                    <Link to={check.url}>
                                        {check.title}{' '}
                                        <span className="font-weight-normal text-muted">#{check.idWithoutKind}</span>
                                    </Link>
                                </th>
                            ))}
                        </tr>
                    </thead>
                    <tbody>
                        {REPOS.map((repo, i) => (
                            <tr key={i}>
                                <th className="text-nowrap" style={{ width: '1%' }}>
                                    <RepoLink repoName={repo} to={`/${repo}`} />
                                </th>
                                {threadsOrError.nodes.map((thread, i) => (
                                    <td key={i} className="p-0 align-middle">
                                        <CheckDashboardCell
                                            {...props}
                                            thread={thread}
                                            repo={repo}
                                            paddingClassName="p-2"
                                        />
                                    </td>
                                ))}
                            </tr>
                        ))}
                    </tbody>
                </table>
            )}
        </div>
    )
}
