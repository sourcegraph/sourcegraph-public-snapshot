import { from, Observable, of, throwError } from 'rxjs'
import { map, mapTo, switchMap, catchError, tap } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { memoizeObservable } from '../../../../shared/src/util/memoizeObservable'
import { storage } from '../../browser/storage'
import { isExtension } from '../../context'
import { resolveRepo } from '../../shared/repo/backend'
import { normalizeRepoName } from './util'
import { ajax } from 'rxjs/ajax'
import { EREPONOTFOUND } from '../../../../shared/src/backend/errors'

interface PhabEntity {
    id: string // e.g. "48"
    type: string // e.g. "RHURI"
    phid: string // e.g. "PHID-RHURI-..."
}

interface ConduitURI extends PhabEntity {
    fields: {
        uri: {
            raw: string // e.g. https://secure.phabricator.com/source/phabricator.git",
            display: string // e.g. https://secure.phabricator.com/source/phabricator.git",
            effective: string // e.g. https://secure.phabricator.com/source/phabricator.git",
            normalized: string // e.g. secure.phabricator.com/source/phabricator",
        }
    }
}

interface ConduitRepo extends PhabEntity {
    fields: {
        name: string
        vcs: string // e.g. 'git'
        callsign: string
        shortName: string
        status: 'active' | 'inactive'
    }
    attachments: {
        uris: {
            uris: ConduitURI[]
        }
    }
}

interface ConduitReposResponse {
    cursor: {
        limit: number
        after: number | null
        before: number | null
    }
    data: ConduitRepo[]
}

interface ConduitRef {
    ref: string
    type: 'base' | 'diff'
    commit: string // a SHA
    remote: {
        uri: string
    }
}

interface ConduitDiffChange {
    oldPath: string
    currentPath: string
}

interface ConduitDiffDetails {
    branch: string
    sourceControlBaseRevision: string // the merge base commit
    description: string // e.g. 'rNZAP9bee3bc2cd3068dd97dfa87068c4431c5d6093ef'
    changes: ConduitDiffChange[]
    dateCreated: string
    authorName: string
    authorEmail: string
    properties: {
        'arc.staging': {
            status: string
            refs: ConduitRef[]
        }
        'local:commits': string[]
    }
}

interface ConduitDiffDetailsResponse {
    [id: string]: ConduitDiffDetails
}

function createConduitRequestForm(): FormData {
    const searchForm = document.querySelector('.phabricator-search-menu form')
    if (!searchForm) {
        throw new Error('cannot create conduit request form')
    }
    const form = new FormData()
    form.set('__csrf__', searchForm.querySelector<HTMLInputElement>('input[name=__csrf__]')!.value)
    form.set('__form__', searchForm.querySelector<HTMLInputElement>('input[name=__form__]')!.value)
    return form
}

/**
 * Native installation of the Phabricator extension does not allow for us to fetch the style.bundle from a script element.
 * To get around this we fetch the bundled CSS contents and append it to the DOM.
 */
export async function getPhabricatorCSS(sourcegraphURL: string): Promise<string> {
    const bundleUID = process.env.BUNDLE_UID
    const resp = await fetch(sourcegraphURL + `/.assets/extension/css/style.bundle.css?v=${bundleUID}`, {
        method: 'GET',
        credentials: 'include',
        headers: new Headers({ Accept: 'text/html' }),
    })
    return resp.text()
}

type ConduitResponse<T> =
    | { error_code: null; error_info: null; result: T }
    | { error_code: string; error_info: string; result: null }

function queryConduit<T>(endpoint: string, params: {}): Observable<T> {
    const form = createConduitRequestForm()
    for (const [key, value] of Object.entries(params)) {
        form.set(`params[${key}]`, JSON.stringify(value))
    }
    return ajax({
        url: window.location.origin + endpoint,
        method: 'POST',
        body: form,
        withCredentials: true,
        headers: {
            Accept: 'application/json',
        },
    }).pipe(
        map(({ response }: { response: ConduitResponse<T> }) => {
            if (response.error_code !== null) {
                throw new Error(`error ${response.error_code}: ${response.error_info}`)
            }
            return response.result
        })
    )
}

function getDiffDetailsFromConduit(diffID: number, differentialID: number): Observable<ConduitDiffDetails> {
    return queryConduit<ConduitDiffDetailsResponse>('/api/differential.querydiffs', {
        ids: [diffID],
        revisionIDs: [differentialID],
    }).pipe(map(diffDetails => diffDetails[`${diffID}`]))
}

function getRawDiffFromConduit(diffID: number): Observable<string> {
    return queryConduit<string>('/api/differential.getrawdiff', { diffID })
}

interface ConduitDifferentialQueryResponse {
    [index: string]: {
        repositoryPHID: string | null
    }
}

function getRepoPHIDForDifferentialID(differentialID: number): Observable<string> {
    return queryConduit<ConduitDifferentialQueryResponse>('/api/differential.query', { ids: [differentialID] }).pipe(
        map(result => {
            const phid = result['0'].repositoryPHID
            if (!phid) {
                // This happens for diffs that were created without an associated repository
                throw new Error(`no repositoryPHID for diff ${differentialID}`)
            }
            return phid
        })
    )
}

interface CreatePhabricatorRepoOptions extends Pick<PlatformContext, 'requestGraphQL'> {
    callsign: string
    repoName: string
    phabricatorURL: string
}

const createPhabricatorRepo = memoizeObservable(
    ({ requestGraphQL, ...variables }: CreatePhabricatorRepoOptions): Observable<void> =>
        requestGraphQL<GQL.IMutation>({
            request: gql`
                mutation addPhabricatorRepo($callsign: String!, $repoName: String!, $phabricatorURL: String!) {
                    addPhabricatorRepo(callsign: $callsign, uri: $repoName, url: $phabricatorURL) {
                        alwaysNil
                    }
                }
            `,
            variables,
            mightContainPrivateInfo: true,
        }).pipe(mapTo(undefined)),
    ({ callsign }) => callsign
)

interface PhabricatorRepoDetails {
    callsign: string
    rawRepoName: string
}

export function getRepoDetailsFromCallsign(
    callsign: string,
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<PhabricatorRepoDetails> {
    return queryConduit<ConduitReposResponse>('/api/diffusion.repository.search', {
        constraints: { callsigns: [callsign] },
        attachments: { uris: true },
    }).pipe(
        switchMap(({ data }) => {
            const repo = data[0]
            if (!repo) {
                throw new Error(`could not locate repo with callsign ${callsign}`)
            }
            if (!repo.attachments || !repo.attachments.uris) {
                throw new Error(`could not locate git uri for repo with callsign ${callsign}`)
            }
            return convertConduitRepoToRepoDetails(repo)
        }),
        switchMap(details => {
            if (!details) {
                return throwError(new Error('could not parse repo details'))
            }
            return createPhabricatorRepo({
                callsign,
                repoName: details.rawRepoName,
                phabricatorURL: window.location.origin,
                requestGraphQL,
            }).pipe(mapTo(details))
        })
    )
}

/**
 * Queries the sourcegraph.configuration conduit API endpoint.
 *
 * The Phabricator extension updates the window object automatically, but in the
 * case it fails we query the conduit API.
 */
export function getSourcegraphURLFromConduit(): Observable<string> {
    return queryConduit<{ url: string }>('/api/sourcegraph.configuration', {}).pipe(map(({ url }) => url))
}

function getRepoDetailsFromRepoPHID(
    phid: string,
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<PhabricatorRepoDetails> {
    const form = createConduitRequestForm()
    form.set('params[constraints]', JSON.stringify({ phids: [phid] }))
    form.set('params[attachments]', '{ "uris": true }')

    return queryConduit<ConduitReposResponse>('/api/diffusion.repository.search', {
        constraints: {
            phids: [phid],
        },
        attachments: {
            uris: true,
        },
    }).pipe(
        switchMap(({ data }) => {
            const repo = data[0]
            if (!repo) {
                throw new Error(`could not locate repo with phid ${phid}`)
            }
            if (!repo.attachments || !repo.attachments.uris) {
                throw new Error(`could not locate git uri for repo with phid ${phid}`)
            }
            return from(convertConduitRepoToRepoDetails(repo)).pipe(
                switchMap(details => {
                    if (!details) {
                        return throwError(new Error('could not parse repo details'))
                    }
                    return createPhabricatorRepo({
                        callsign: repo.fields.callsign,
                        repoName: details.rawRepoName,
                        phabricatorURL: window.location.origin,
                        requestGraphQL,
                    }).pipe(mapTo(details))
                })
            )
        })
    )
}

export function getRepoDetailsFromDifferentialID(
    differentialID: number,
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<PhabricatorRepoDetails> {
    return getRepoPHIDForDifferentialID(differentialID).pipe(
        switchMap(repositoryPHID => getRepoDetailsFromRepoPHID(repositoryPHID, requestGraphQL))
    )
}

async function convertConduitRepoToRepoDetails(repo: ConduitRepo): Promise<PhabricatorRepoDetails | null> {
    if (isExtension) {
        const items = await storage.sync.get()
        if (items.phabricatorMappings) {
            for (const mapping of items.phabricatorMappings) {
                if (mapping.callsign === repo.fields.callsign) {
                    return {
                        callsign: repo.fields.callsign,
                        rawRepoName: mapping.path,
                    }
                }
            }
        }
        return convertToDetails(repo)
    }
    // The path to a phabricator repository on a Sourcegraph instance may differ than it's URI / name from the
    // phabricator conduit API. Since we do not currently send the PHID with the Phabricator repository this a
    // backwards work around configuration setting to ensure mappings are correct. This logic currently exists
    // in the browser extension options menu.
    type Mappings = { callsign: string; path: string }[]
    const mappingsString = window.localStorage.getItem('PHABRICATOR_CALLSIGN_MAPPINGS')
    const callsignMappings = mappingsString
        ? (JSON.parse(mappingsString) as Mappings)
        : window.PHABRICATOR_CALLSIGN_MAPPINGS || []
    const details = convertToDetails(repo)
    if (callsignMappings) {
        for (const mapping of callsignMappings) {
            if (mapping.callsign === repo.fields.callsign) {
                return {
                    callsign: repo.fields.callsign,
                    rawRepoName: mapping.path,
                }
            }
        }
    }
    return details
}

function convertToDetails(repo: ConduitRepo): PhabricatorRepoDetails | null {
    let uri: ConduitURI | undefined
    for (const u of repo.attachments.uris.uris) {
        const normalPath = u.fields.uri.normalized.replace('\\', '')
        if (normalPath.startsWith(window.location.host + '/')) {
            continue
        }
        uri = u
        break
    }
    if (!uri) {
        return null
    }
    const rawURI = uri.fields.uri.raw
    const rawRepoName = normalizeRepoName(rawURI)
    return { callsign: repo.fields.callsign, rawRepoName }
}

interface ResolveStagingOptions extends Pick<PlatformContext, 'requestGraphQL'> {
    repoName: string
    diffID: number
    baseRev: string
    patch?: string
    date?: string
    authorName?: string
    authorEmail?: string
    description?: string
}

const resolveStagingRev = memoizeObservable(
    ({ requestGraphQL, ...variables }: ResolveStagingOptions): Observable<string | null> =>
        requestGraphQL<GQL.IMutation>({
            request: gql`
                mutation ResolveStagingRev(
                    $repoName: String!
                    $diffID: ID!
                    $baseRev: String!
                    $patch: String
                    $date: String
                    $authorName: String
                    $authorEmail: String
                    $description: String
                ) {
                    resolvePhabricatorDiff(
                        repoName: $repoName
                        diffID: $diffID
                        baseRev: $baseRev
                        patch: $patch
                        date: $date
                        authorName: $authorName
                        authorEmail: $authorEmail
                        description: $description
                    ) {
                        oid
                    }
                }
            `,
            variables,
            mightContainPrivateInfo: true,
        }).pipe(
            map(dataOrThrowErrors),
            map(result => {
                if (!result.resolvePhabricatorDiff) {
                    throw new Error('Empty resolvePhabricatorDiff')
                }
                return result.resolvePhabricatorDiff.oid
            })
        ),
    ({ diffID }: ResolveStagingOptions) => diffID.toString()
)

function hasThisFileChanged(filePath: string, changes: ConduitDiffChange[]): boolean {
    for (const change of changes) {
        if (change.currentPath === filePath) {
            return true
        }
    }
    return false
}

interface ResolveDiffOpt {
    repoName: string
    filePath: string
    differentialID: number
    diffID: number
    leftDiffID?: number
    isBase: boolean
    useDiffForBase: boolean // indicates whether the base should use the diff commit
    useBaseForDiff: boolean // indicates whether the diff should use the base commit
}

interface PropsWithInfo extends ResolveDiffOpt {
    info: ConduitDiffDetails
}

function getPropsWithInfo(props: ResolveDiffOpt): Observable<PropsWithInfo> {
    return getDiffDetailsFromConduit(props.diffID, props.differentialID).pipe(
        switchMap(info => {
            if (props.isBase || !props.leftDiffID || hasThisFileChanged(props.filePath, info.changes)) {
                // no need to update props
                return of({
                    ...props,
                    info,
                })
            }
            return getDiffDetailsFromConduit(props.leftDiffID, props.differentialID).pipe(
                map(
                    (info: ConduitDiffDetails): PropsWithInfo => ({
                        ...props,
                        info,
                        diffID: props.leftDiffID!,
                        useBaseForDiff: true,
                    })
                )
            )
        })
    )
}

function getStagingDetails(
    propsWithInfo: PropsWithInfo
): { repoName: string; ref: ConduitRef; unconfigured: boolean } | undefined {
    const stagingInfo = propsWithInfo.info.properties['arc.staging']
    if (!stagingInfo) {
        return undefined
    }
    let key: string
    if (propsWithInfo.isBase) {
        const type = propsWithInfo.useDiffForBase ? 'diff' : 'base'
        key = `refs/tags/phabricator/${type}/${propsWithInfo.diffID}`
    } else {
        const type = propsWithInfo.useBaseForDiff ? 'base' : 'diff'
        key = `refs/tags/phabricator/${type}/${propsWithInfo.diffID}`
    }
    for (const ref of propsWithInfo.info.properties['arc.staging'].refs) {
        if (ref.ref === key) {
            const remote = ref.remote.uri
            if (remote) {
                return {
                    repoName: normalizeRepoName(remote),
                    ref,
                    unconfigured: stagingInfo.status === 'repository.unconfigured',
                }
            }
        }
    }
    return undefined
}

interface ResolvedDiff {
    commitID: string
    stagingRepoName?: string
}

export function resolveDiffRev(
    props: ResolveDiffOpt,
    requestGraphQL: PlatformContext['requestGraphQL'],
    part: 'base' | 'head'
): Observable<ResolvedDiff> {
    console.log(`Resolving diff rev for ${part}`, props)
    const ensureCommitID = (commitID: string | null): { commitID: string } => {
        if (!commitID) {
            throw new Error('commitID cannot be null')
        }
        return { commitID }
    }
    return getPropsWithInfo(props).pipe(
        switchMap(propsWithInfo => {
            const stagingDetails = getStagingDetails(propsWithInfo)
            const conduitProps = {
                repoName: propsWithInfo.repoName,
                diffID: propsWithInfo.diffID,
                baseRev: propsWithInfo.info.sourceControlBaseRevision,
                date: propsWithInfo.info.dateCreated,
                authorName: propsWithInfo.info.authorName,
                authorEmail: propsWithInfo.info.authorEmail,
                description: propsWithInfo.info.description,
            }

            // When resolving the base, the commit ID is found directly on the diff description.
            if (propsWithInfo.isBase && !propsWithInfo.useDiffForBase) {
                return of({ commitID: propsWithInfo.info.sourceControlBaseRevision })
            }
            if (!stagingDetails || stagingDetails.unconfigured) {
                // If there are no staging details, get the patch from the conduit API,
                // create a one-off commit on the Sourcegraph instance from the patch,
                // and resolve to the commit ID returned by the Sourcegraph instance.
                return getRawDiffFromConduit(propsWithInfo.diffID).pipe(
                    switchMap(patch => resolveStagingRev({ ...conduitProps, patch, requestGraphQL })),
                    map(ensureCommitID)
                )
            }

            // If staging details are configured, first check if the repo is present on the Sourcegraph instance.
            return resolveRepo({ rawRepoName: stagingDetails.repoName, requestGraphQL }).pipe(
                // If the repo is present on the Sourcegraph instance,
                // use the commitID and repo name from the staging details.
                mapTo({
                    commitID: stagingDetails.ref.commit,
                    stagingRepoName: stagingDetails.repoName,
                }),
                // Otherwise, create a one-off commit containing the patch on the Sourcegraph instance,
                // and resolve to the commit ID returned by the Sourcegraph instance.
                catchError(error => {
                    if (error.code !== EREPONOTFOUND) {
                        throw error
                    }
                    return getRawDiffFromConduit(propsWithInfo.diffID).pipe(
                        switchMap(patch => resolveStagingRev({ ...conduitProps, patch, requestGraphQL })),
                        map(ensureCommitID)
                    )
                })
            )
        })
    )
}
