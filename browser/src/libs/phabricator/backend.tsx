import { from, Observable, of } from 'rxjs'
import { map, mapTo, switchMap, catchError, tap } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { memoizeObservable } from '../../../../shared/src/util/memoizeObservable'
import { storage } from '../../browser/storage'
import { isExtension } from '../../context'
import { resolveRepo } from '../../shared/repo/backend'
import { normalizeRepoName } from './util'

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
    error_code?: string
    error_info?: string
    result: {
        cursor: {
            limit: number
            after: number | null
            before: number | null
        }
        data: ConduitRepo[]
    }
}

interface SourcegraphConduitConfiguration {
    result: {
        url: string
    }
    error_code?: string
    error_info?: string
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
    error_code?: string
    error_info?: string
    result: {
        [id: string]: ConduitDiffDetails
    }
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

async function getDiffDetailsFromConduit(diffID: number, differentialID: number): Promise<ConduitDiffDetails> {
    const form = createConduitRequestForm()
    form.set('params[ids]', `[${diffID}]`)
    form.set('params[revisionIDs]', `[${differentialID}]`)

    const res = await (await fetch(window.location.origin + '/api/differential.querydiffs', {
        method: 'POST',
        body: form,
        credentials: 'include',
        headers: new Headers({ Accept: 'application/json' }),
    })).json()
    if (res.error_code) {
        throw new Error(`error ${res.error_code}: ${res.error_info}`)
    }
    return res.result['' + diffID]
}

interface ConduitRawDiffResponse {
    error_code?: string
    error_info?: string
    result: string
}

function getRawDiffFromConduit(diffID: number): Promise<string> {
    return new Promise((resolve, reject) => {
        const form = createConduitRequestForm()
        form.set('params[diffID]', diffID.toString())

        fetch(window.location.origin + '/api/differential.getrawdiff', {
            method: 'POST',
            body: form,
            credentials: 'include',
            headers: new Headers({ Accept: 'application/json' }),
        })
            .then(resp => resp.json())
            .then((res: ConduitRawDiffResponse) => {
                if (res.error_code) {
                    reject(new Error(`error ${res.error_code}: ${res.error_info}`))
                }
                resolve(res.result)
            })
            .catch(reject)
    })
}

interface ConduitCommit {
    fields: {
        identifier: string
    }
}

interface ConduitDiffusionCommitQueryResponse {
    error_code?: string
    error_info?: string
    result: {
        data: ConduitCommit[]
    }
}

function searchForCommitID(props: any): Promise<string> {
    return new Promise((resolve, reject) => {
        const form = createConduitRequestForm()
        form.set('params[constraints]', `{"ids":[${props.diffID}]}`)

        fetch(window.location.origin + '/api/diffusion.commit.search', {
            method: 'POST',
            body: form,
            credentials: 'include',
            headers: new Headers({ Accept: 'application/json' }),
        })
            .then(resp => resp.json())
            .then((resp: ConduitDiffusionCommitQueryResponse) => {
                if (resp.error_code) {
                    reject(new Error(`error ${resp.error_code}: ${resp.error_info}`))
                }

                resolve(resp.result.data[0].fields.identifier)
            })
            .catch(reject)
    })
}

interface ConduitDifferentialQueryResponse {
    error_code?: string
    error_info?: string
    result: {
        [index: string]: {
            repositoryPHID: string | null
        }
    }
}

async function getRepoPHIDForDifferentialID(differentialID: number): Promise<string> {
    const form = createConduitRequestForm()
    form.set('params[ids]', `[${differentialID}]`)
    const response = await fetch(window.location.origin + '/api/differential.query', {
        method: 'POST',
        body: form,
        credentials: 'include',
        headers: new Headers({ Accept: 'application/json' }),
    })
    const responseJSON: ConduitDifferentialQueryResponse = await response.json()
    if (responseJSON.error_code) {
        throw new Error(`error ${responseJSON.error_code}: ${responseJSON.error_info}`)
    }
    const phid = responseJSON.result['0'].repositoryPHID
    if (!phid) {
        // This happens for diffs that were created without an associated repository
        throw new Error(`no repositoryPHID for diff ${differentialID}`)
    }
    return phid
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
): Promise<PhabricatorRepoDetails> {
    return new Promise((resolve, reject) => {
        const form = createConduitRequestForm()
        form.set('params[constraints]', JSON.stringify({ callsigns: [callsign] }))
        form.set('params[attachments]', '{ "uris": true }')
        fetch(window.location.origin + '/api/diffusion.repository.search', {
            method: 'POST',
            body: form,
            credentials: 'include',
            headers: new Headers({ Accept: 'application/json' }),
        })
            .then(resp => resp.json())
            .then((res: ConduitReposResponse) => {
                if (res.error_code) {
                    reject(new Error(`error ${res.error_code}: ${res.error_info}`))
                }
                const repo = res.result.data[0]
                if (!repo) {
                    reject(new Error(`could not locate repo with callsign ${callsign}`))
                }
                if (!repo.attachments || !repo.attachments.uris) {
                    reject(new Error(`could not locate git uri for repo with callsign ${callsign}`))
                }

                return convertConduitRepoToRepoDetails(repo).then(details => {
                    if (details) {
                        return createPhabricatorRepo({
                            callsign,
                            repoName: details.rawRepoName,
                            phabricatorURL: window.location.origin,
                            requestGraphQL,
                        }).subscribe(() => resolve(details))
                    }
                    return reject(new Error('could not parse repo details'))
                })
            })
            .catch(reject)
    })
}

/**
 * Queries the sourcegraph.configuration conduit API endpoint.
 *
 * The Phabricator extension updates the window object automatically, but in the
 * case it fails we query the conduit API.
 */
export async function getSourcegraphURLFromConduit(): Promise<string> {
    const form = createConduitRequestForm()
    const res: SourcegraphConduitConfiguration = await fetch(
        window.location.origin + '/api/sourcegraph.configuration',
        {
            method: 'POST',
            body: form,
            credentials: 'include',
            headers: new Headers({ Accept: 'application/json' }),
        }
    ).then(resp => resp.json())
    if (res.error_code) {
        throw new Error(`error ${res.error_code}: ${res.error_info}`)
    }
    if (!res || !res.result) {
        throw new Error(`error ${res}. could not fetch sourcegraph configuration.`)
    }
    return res.result.url
}

function getRepoDetailsFromRepoPHID(
    phid: string,
    requestGraphQL: PlatformContext['requestGraphQL']
): Promise<PhabricatorRepoDetails> {
    return new Promise((resolve, reject) => {
        const form = createConduitRequestForm()
        form.set('params[constraints]', JSON.stringify({ phids: [phid] }))
        form.set('params[attachments]', '{ "uris": true }')

        fetch(window.location.origin + '/api/diffusion.repository.search', {
            method: 'POST',
            body: form,
            credentials: 'include',
            headers: new Headers({ Accept: 'application/json' }),
        })
            .then(resp => resp.json())
            .then((res: ConduitReposResponse) => {
                if (res.error_code) {
                    throw new Error(`error ${res.error_code}: ${res.error_info}`)
                }
                const repo = res.result.data[0]
                if (!repo) {
                    throw new Error(`could not locate repo with phid ${phid}`)
                }
                if (!repo.attachments || !repo.attachments.uris) {
                    throw new Error(`could not locate git uri for repo with phid ${phid}`)
                }

                return convertConduitRepoToRepoDetails(repo).then(details => {
                    if (details) {
                        return createPhabricatorRepo({
                            callsign: repo.fields.callsign,
                            repoName: details.rawRepoName,
                            phabricatorURL: window.location.origin,
                            requestGraphQL,
                        })
                            .pipe(map(() => details))
                            .subscribe(() => {
                                resolve(details)
                            })
                    }
                    return reject(new Error('could not parse repo details'))
                })
            })
            .catch(reject)
    })
}

export async function getRepoDetailsFromDifferentialID(
    differentialID: number,
    requestGraphQL: PlatformContext['requestGraphQL']
): Promise<PhabricatorRepoDetails> {
    const repositoryPHID = await getRepoPHIDForDifferentialID(differentialID)
    return await getRepoDetailsFromRepoPHID(repositoryPHID, requestGraphQL)
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

async function getPropsWithInfo(props: ResolveDiffOpt): Promise<PropsWithInfo> {
    let info = await getDiffDetailsFromConduit(props.diffID, props.differentialID)
    if (props.isBase || !props.leftDiffID || hasThisFileChanged(props.filePath, info.changes)) {
        // no need to update props
        return {
            ...props,
            info,
        }
    }
    info = await getDiffDetailsFromConduit(props.leftDiffID, props.differentialID)
    return {
        ...props,
        info,
        diffID: props.leftDiffID!,
        useBaseForDiff: true,
    }
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
    return from(getPropsWithInfo(props)).pipe(
        switchMap(propsWithInfo => {
            console.log('propsWithInfo', propsWithInfo)
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
            if (propsWithInfo.isBase && !propsWithInfo.useDiffForBase) {
                console.log(1)
                return of({ commitID: propsWithInfo.info.sourceControlBaseRevision })
            }
            if (!stagingDetails || stagingDetails.unconfigured) {
                console.log('No staging details or staging details unconfigured')
                // The last diff (final commit) is not found in the staging area, but rather on the description.
                console.log(2)
                return from(getRawDiffFromConduit(propsWithInfo.diffID)).pipe(
                    switchMap(patch => resolveStagingRev({ ...conduitProps, patch, requestGraphQL })),
                    map(ensureCommitID)
                )
            }

            if (!stagingDetails.unconfigured) {
                console.log(3, stagingDetails)
                // Ensure the staging repo exists before resolving. Otherwise create the patch.
                return resolveRepo({ rawRepoName: stagingDetails.repoName, requestGraphQL }).pipe(
                    tap(() => {
                        console.log('resolvedRepo')
                    }),
                    mapTo({
                        commitID: stagingDetails.ref.commit,
                        stagingRepoName: stagingDetails.repoName,
                    }),
                    catchError(error => {
                        console.log('3 error', error)
                        return from(getRawDiffFromConduit(propsWithInfo.diffID)).pipe(
                            switchMap(patch => resolveStagingRev({ ...conduitProps, patch, requestGraphQL })),
                            map(ensureCommitID),
                            tap(({ commitID }) => {
                                console.log(4, commitID, part)
                            })
                        )
                    })
                )
            }

            if (!propsWithInfo.isBase) {
                for (const commit of Object.keys(propsWithInfo.info.properties['local:commits'])) {
                    console.log(5)
                    return of({ commitID: commit })
                }
            }
            // last ditch effort to search conduit API for commit ID
            console.log(6)
            return from(searchForCommitID(propsWithInfo)).pipe(map(commitID => ({ commitID })))
        })
    )
}
