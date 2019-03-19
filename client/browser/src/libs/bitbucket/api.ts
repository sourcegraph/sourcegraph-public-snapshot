import { isDefined } from '@sourcegraph/codeintellify/lib/helpers'
import { first } from 'lodash'
import { Observable } from 'rxjs'
import { ajax } from 'rxjs/ajax'
import { filter, map } from 'rxjs/operators'

import { memoizeObservable } from '../../../../../shared/src/util/memoizeObservable'
import { PRPageInfo } from './scrape'

//
// PR API /rest/api/1.0/projects/SG/repos/go-langserver/pull-requests/1

const buildURL = (project: string, repoSlug: string, path: string) =>
    `${window.location.origin}/rest/api/1.0/projects/${encodeURIComponent(project)}/repos/${repoSlug}${path}`

const get = <T>(url: string): Observable<T> => ajax.get(url).pipe(map(({ response }) => response as T))

interface Repo {
    project: { key: string }
    name: string
    public: boolean
}

interface Ref {
    /**
     * The branch name.
     */
    displayId: string
    /**
     * The commit ID.
     */
    latestCommit: string

    repository: Repo
}

interface PRResponse {
    fromRef: Ref
    toRef: Ref
}

interface GetCommitsForPRInput extends Pick<PRPageInfo, 'project' | 'repoSlug'> {
    /** Required here. */
    prID: number
}

/**
 * Get the base commit ID for a merge request.
 */
export const getCommitsForPR: (
    info: GetCommitsForPRInput
) => Observable<{ baseCommitID: string; headCommitID: string }> = memoizeObservable(
    ({ project, repoSlug, prID }) =>
        get<PRResponse>(buildURL(project, repoSlug, `/pull-requests/${prID}`)).pipe(
            map(({ fromRef, toRef }) => ({ baseCommitID: toRef.latestCommit, headCommitID: fromRef.latestCommit }))
        ),
    ({ prID }) => prID.toString()
)

interface GetBaseCommitInput extends Pick<PRPageInfo, 'project' | 'repoSlug'> {
    commitID: string
}

interface Commit {
    id: string
}

interface CommitResponse {
    parents: Commit[]
}

// Commit API /rest/api/1.0/projects/SG/repos/go-langserver/commits/b8a948dc75cc9d0c01ece01d0ba9d1eeace573aa
export const getBaseCommit: (info: GetBaseCommitInput) => Observable<string> = memoizeObservable(
    ({ project, repoSlug, commitID }) =>
        get<CommitResponse>(buildURL(project, repoSlug, `/commits/${commitID}`)).pipe(
            map(({ parents }) => first(parents)),
            filter(isDefined),
            map(({ id }) => id)
        )
)
