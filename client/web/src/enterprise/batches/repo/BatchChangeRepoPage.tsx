import * as H from 'history'
import React, { ReactElement } from 'react'

import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { PageHeader } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'
import { RepositoryFields } from '../../../graphql-operations'
import { queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs } from '../detail/backend'
import { BatchChangeStatsTotalAction } from '../detail/BatchChangeStatsCard'
import {
    ChangesetStatusUnpublished,
    ChangesetStatusOpen,
    ChangesetStatusClosed,
    ChangesetStatusMerged,
} from '../detail/changesets/ChangesetStatusCell'

import { queryRepoBatchChanges as _queryRepoBatchChanges } from './backend'
import { RepoBatchChanges } from './RepoBatchChanges'

interface BatchChangeRepoPageProps extends ThemeProps {
    history: H.History
    location: H.Location
    repo: RepositoryFields
    /** For testing only. */
    queryRepoBatchChanges?: typeof _queryRepoBatchChanges
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
}

export const BatchChangeRepoPage: React.FunctionComponent<BatchChangeRepoPageProps> = ({ repo, ...context }) => {
    const repoDisplayName = displayRepoName(repo.name)

    return (
        <Page>
            <PageTitle title="Batch Changes" />
            <PageHeader path={[{ icon: BatchChangesIcon, text: 'Batch Changes' }]} headingElement="h1" />
            <div className="d-flex align-items-center mt-4 mb-3">
                <h2 className="mb-0">{repoDisplayName}</h2>
                <div className="d-flex flex-1 ml-2">+1000 •4000 -2000</div>
                <StatsBar />
            </div>
            <p>Batch changes has created 78 changesets on {repoDisplayName}</p>
            <RepoBatchChanges viewerCanAdminister={true} repo={repo} {...context} />
        </Page>
    )
}

const ACTION_CLASSNAMES = 'd-flex flex-column text-muted justify-content-center align-items-center mx-2'

// TODO: Generalize icon label type to accept strings
const element = (string: string): ReactElement => <span>{string}</span>

const StatsBar: React.FunctionComponent<{}> = () => (
    <div className="d-flex flex-wrap align-items-center">
        <BatchChangeStatsTotalAction count={78} />
        <ChangesetStatusOpen className={ACTION_CLASSNAMES} label={element(`${3} Open`)} />
        <ChangesetStatusUnpublished className={ACTION_CLASSNAMES} label={element(`${1} Unpublished`)} />
        <ChangesetStatusClosed className={ACTION_CLASSNAMES} label={element(`${5} Closed`)} />
        <ChangesetStatusMerged className={ACTION_CLASSNAMES} label={element(`${67} Merged`)} />
    </div>
)
