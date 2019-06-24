import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useState } from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { HeroPage } from '../../../components/HeroPage'
import { WithQueryParameter } from '../../../components/withQueryParameter/WithQueryParameter'
import { createThreadAreaContext } from '../../threads/detail/ThreadArea'
import { useChangesetByID } from '../components/useChangesetByID'
import { ChangesetFilesList } from '../detail/changes/ChangesetFilesList'
import { ChangesetsAreaContext } from '../global/ChangesetsArea'
import { ChangesetIcon } from '../icons'
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
    if (threadOrError === LOADING) {
        return null // loading
    }
    if (isErrorLike(threadOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={threadOrError.message} />
    }
    const context = createThreadAreaContext(props, { thread: threadOrError, onThreadUpdate: setThreadOrError })

    const isCreateFormExpanded = new URLSearchParams(props.location.search).get(CREATE_FORM_EXPANDED_PARAM) !== null

    return (
        <div className="changeset-preview-page mt-3 overflow-auto">
            <div className="container">
                <h1 className="mb-3">Preview changeset</h1>
                {isCreateFormExpanded ? (
                    <>
                        <CreateChangesetFromPreviewForm {...context} className="border p-3 mb-6" />
                        <div className="my-6" />
                    </>
                ) : (
                    <div className="alert alert-warning d-flex align-items-center mb-3">
                        <Link to={CREATE_FORM_EXPANDED_URL} className="btn btn-lg btn-success mr-3">
                            <ChangesetIcon className="icon-inline mr-1" /> Create changeset
                        </Link>
                        <span className="text-muted">
                            Create branches for this change in all affected repositories and request code reviews.
                        </span>
                    </div>
                )}
                <ChangesetSummaryBar {...context} />
            </div>
            <hr className="my-4" />
            <div className="container">
                <WithQueryParameter defaultQuery="" history={props.history} location={props.location}>
                    {({ query, onQueryChange }) => (
                        <ChangesetFilesList {...props} {...context} query={query} onQueryChange={onQueryChange} />
                    )}
                </WithQueryParameter>
            </div>
        </div>
    )
}
