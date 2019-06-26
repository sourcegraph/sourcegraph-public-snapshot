import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useState } from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { HeroPage } from '../../../components/HeroPage'
import { WithQueryParameter } from '../../../components/withQueryParameter/WithQueryParameter'
import { createThreadAreaContext } from '../../threads/detail/ThreadArea'
import { ChangesetActionsList } from '../detail/changes/ChangesetActionsList'
import { ChangesetCommitsList } from '../detail/changes/ChangesetCommitsList'
import { ChangesetFilesList } from '../detail/changes/ChangesetFilesList'
import { ChangesetRepositoriesList } from '../detail/changes/ChangesetRepositoriesList'
import { ChangesetTasksList } from '../detail/changes/ChangesetTasksList'
import { ChangesetsAreaContext } from '../global/ChangesetsArea'
import { ChangesetIcon } from '../icons'
import { useChangesetByID } from '../util/useChangesetByID'
import { useExtraChangesetInfo } from '../util/useExtraChangesetInfo'
import { ChangesetSummaryBar } from './ChangesetSummaryBar'
import { CreateChangesetFromPreviewForm } from './CreateChangesetFromPreviewForm'

interface Props extends ChangesetsAreaContext, RouteComponentProps<{ threadID: string }> {}

const LOADING: 'loading' = 'loading'

const CREATE_FORM_EXPANDED_PARAM = 'expand'
const CREATE_FORM_EXPANDED_URL: H.LocationDescriptor = {
    search: new URLSearchParams({ [CREATE_FORM_EXPANDED_PARAM]: '1' }).toString(),
}

/**
 * A page that shows a preview of a changeset created from code actions.
 */
export const ChangesetPreviewPage: React.FunctionComponent<Props> = props => {
    const [threadOrError, setThreadOrError] = useChangesetByID(props.match.params.threadID)
    const xchangeset = useExtraChangesetInfo(threadOrError)
    if (threadOrError === LOADING || xchangeset === LOADING) {
        return null // loading
    }
    if (isErrorLike(threadOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={threadOrError.message} />
    }
    if (isErrorLike(xchangeset)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={xchangeset.message} />
    }
    const context = createThreadAreaContext(props, { thread: threadOrError, onThreadUpdate: setThreadOrError })

    const isCreateFormExpanded = new URLSearchParams(props.location.search).get(CREATE_FORM_EXPANDED_PARAM) !== null

    return (
        <div className="changeset-preview-page mt-3 overflow-auto">
            <div className="container">
                <h1 className="mb-3">Preview changeset</h1>
                {isCreateFormExpanded ? (
                    <CreateChangesetFromPreviewForm {...context} className="border p-3 mb-4" history={props.history} />
                ) : (
                    <div className="alert alert-warning d-flex align-items-center mb-4">
                        <Link to={CREATE_FORM_EXPANDED_URL} className="btn btn-lg btn-success mr-3">
                            <ChangesetIcon className="icon-inline mr-1" /> Create changeset
                        </Link>
                        <span className="text-muted">
                            Create branches for this change in all affected repositories and request code reviews.
                        </span>
                    </div>
                )}
                <ChangesetSummaryBar {...context} xchangeset={xchangeset} />
            </div>
            <hr className="my-4" />
            <div className="container">
                <ChangesetActionsList {...props} {...context} xchangeset={xchangeset} />
                <ChangesetRepositoriesList {...props} {...context} xchangeset={xchangeset} showCommits={true} />
                <ChangesetCommitsList {...props} {...context} xchangeset={xchangeset} className="d-none" />
                <ChangesetTasksList {...props} {...context} xchangeset={xchangeset} />
                <WithQueryParameter defaultQuery="" history={props.history} location={props.location}>
                    {({ query, onQueryChange }) => (
                        <ChangesetFilesList
                            {...props}
                            {...context}
                            xchangeset={xchangeset}
                            query={query}
                            onQueryChange={onQueryChange}
                        />
                    )}
                </WithQueryParameter>
            </div>
        </div>
    )
}
