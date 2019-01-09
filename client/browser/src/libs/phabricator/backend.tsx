import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import storage from '../../browser/storage'
import { isExtension } from '../../context'
import { getContext } from '../../shared/backend/context'
import { mutateGraphQL } from '../../shared/backend/graphql'
import { resolveRepo } from '../../shared/repo/backend'
import { DEFAULT_SOURCEGRAPH_URL, sourcegraphUrl } from '../../shared/util/context'
import { memoizeObservable } from '../../shared/util/memoize'
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
    const searchForm = document.querySelector('.phabricator-search-menu form') as any
    if (!searchForm) {
        throw new Error('cannot create conduit request form')
    }
    const form = new FormData()
    form.set('__csrf__', searchForm.querySelector('input[name=__csrf__]')!.value)
    form.set('__form__', searchForm.querySelector('input[name=__form__]')!.value)
    return form
}

/**
 * Native installation of the Phabricator extension does not allow for us to fetch the style.bundle from a script element.
 * To get around this we fetch the bundled CSS contents and append it to the DOM.
 */
export function getPhabricatorCSS(): Promise<string> {
    const bundleUID = process.env.BUNDLE_UID
    return new Promise((resolve, reject) => {
        fetch(sourcegraphUrl + `/.assets/extension/css/style.bundle.css?v=${bundleUID}`, {
            method: 'GET',
            credentials: 'include',
            headers: new Headers({ Accept: 'text/html' }),
        })
            .then(resp => resp.text())
            .then(resolve)
            .catch(reject)
    })
}

function getDiffDetailsFromConduit(diffID: number, differentialID: number): Promise<ConduitDiffDetails> {
    return new Promise((resolve, reject) => {
        const form = createConduitRequestForm()
        form.set('params[ids]', `[${diffID}]`)
        form.set('params[revisionIDs]', `[${differentialID}]`)

        fetch(window.location.origin + '/api/differential.querydiffs', {
            method: 'POST',
            body: form,
            credentials: 'include',
            headers: new Headers({ Accept: 'application/json' }),
        })
            .then(resp => resp.json())
            .then((res: ConduitDiffDetailsResponse) => {
                if (res.error_code) {
                    reject(new Error(`error ${res.error_code}: ${res.error_info}`))
                }
                resolve(res.result['' + diffID])
            })
            .catch(reject)
    })
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
            // arrays
            repositoryPHID: string
        }
    }
}

function getRepoPHIDForDifferentialID(differentialID: number): Promise<string> {
    return new Promise((resolve, reject) => {
        const form = createConduitRequestForm()
        form.set('params[ids]', `[${differentialID}]`)

        fetch(window.location.origin + '/api/differential.query', {
            method: 'POST',
            body: form,
            credentials: 'include',
            headers: new Headers({ Accept: 'application/json' }),
        })
            .then(resp => resp.json())
            .then((res: ConduitDifferentialQueryResponse) => {
                if (res.error_code) {
                    reject(new Error(`error ${res.error_code}: ${res.error_info}`))
                }
                resolve(res.result['0'].repositoryPHID)
            })
            .catch(reject)
    })
}

interface CreatePhabricatorRepoOptions {
    callsign: string
    repoName: string
    phabricatorURL: string
}

const createPhabricatorRepo = memoizeObservable(
    (options: CreatePhabricatorRepoOptions): Observable<void> =>
        mutateGraphQL({
            ctx: getContext({ repoKey: options.repoName, blacklist: [DEFAULT_SOURCEGRAPH_URL] }),
            request: `mutation addPhabricatorRepo(
            $callsign: String!,
            $repoName: String!
            $phabricatorURL: String!
        ) {
            addPhabricatorRepo(callsign: $callsign, uri: $repoName, url: $phabricatorURL) { alwaysNil }
        }`,
            variables: options,
            retry: false,
        }).pipe(
            map(({ data, errors }) => {
                if (!data || (errors && errors.length > 0)) {
                    throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
                }
            })
        ),
    ({ callsign }) => callsign
)

interface PhabricatorRepoDetails {
    callsign: string
    repoName: string
}

export function getRepoDetailsFromCallsign(callsign: string): Promise<PhabricatorRepoDetails> {
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
                            repoName: details.repoName,
                            phabricatorURL: window.location.origin,
                        }).subscribe(() => resolve(details))
                    } else {
                        reject(new Error('could not parse repo details'))
                    }
                })
            })
            .catch(reject)
    })
}

/**
 *  getSourcegraphURLFromConduit returns the current Sourcegraph URL on the window object or will query the
 *  sourcegraph.configuration conduit API endpoint. The Phabricator extension updates the window object automatically, but in the case it fails
 *  we query the conduit API.
 */
export function getSourcegraphURLFromConduit(): Promise<string> {
    return new Promise((resolve, reject) => {
        const url = window.localStorage.getItem('SOURCEGRAPH_URL') || window.SOURCEGRAPH_URL
        if (url) {
            return resolve(url)
        }
        const form = createConduitRequestForm()
        fetch(window.location.origin + '/api/sourcegraph.configuration', {
            method: 'POST',
            body: form,
            credentials: 'include',
            headers: new Headers({ Accept: 'application/json' }),
        })
            .then(resp => resp.json())
            .then((res: SourcegraphConduitConfiguration) => {
                if (res.error_code) {
                    throw new Error(`error ${res.error_code}: ${res.error_info}`)
                }

                if (!res || !res.result) {
                    throw new Error(`error ${res}. could not fetch sourcegraph configuration.`)
                }
                resolve(res.result.url)
            })
            .catch(reject)
    })
}

function getRepoDetailsFromRepoPHID(phid: string): Promise<PhabricatorRepoDetails> {
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
                            repoName: details.repoName,
                            phabricatorURL: window.location.origin,
                        })
                            .pipe(map(() => details))
                            .subscribe(() => {
                                resolve(details)
                            })
                    } else {
                        reject(new Error('could not parse repo details'))
                    }
                })
            })
            .catch(reject)
    })
}

export function getRepoDetailsFromDifferentialID(differentialID: number): Promise<PhabricatorRepoDetails> {
    return getRepoPHIDForDifferentialID(differentialID).then(getRepoDetailsFromRepoPHID)
}

function convertConduitRepoToRepoDetails(repo: ConduitRepo): Promise<PhabricatorRepoDetails | null> {
    return new Promise((resolve, reject) => {
        if (isExtension) {
            return storage.getSync(items => {
                if (items.phabricatorMappings) {
                    for (const mapping of items.phabricatorMappings) {
                        if (mapping.callsign === repo.fields.callsign) {
                            return resolve({
                                callsign: repo.fields.callsign,
                                repoName: mapping.path,
                            })
                        }
                    }
                }
                return resolve(convertToDetails(repo))
            })
        } else {
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
                        return resolve({
                            callsign: repo.fields.callsign,
                            repoName: mapping.path,
                        })
                    }
                }
            }
            return resolve(details)
        }
    })
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
    const repoName = normalizeRepoName(rawURI)
    return { callsign: repo.fields.callsign, repoName }
}

interface ResolveStagingOptions {
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
    (options: ResolveStagingOptions): Observable<string | null> =>
        mutateGraphQL({
            ctx: getContext({ repoKey: options.repoName, blacklist: [DEFAULT_SOURCEGRAPH_URL] }),
            request: `mutation ResolveStagingRev(
                $repoName: String!,
                $diffID: ID!,
                $baseRev: String!,
                $patch: String,
                $date: String,
                $authorName: String,
                $authorEmail: String,
                $description: String
             ) {
               resolvePhabricatorDiff(
                   repoName: $repoName,
                   diffID: $diffID,
                   baseRev: $baseRev,
                   patch: $patch,
                   date: $date,
                   authorName: $authorName,
                   authorEmail: $authorEmail,
                   description: $description,
               ) {
                   oid
               }
            }`,
            variables: options,
        }).pipe(
            map(({ data, errors }) => {
                if (!(data && data.resolvePhabricatorDiff) || (errors && errors.length > 0)) {
                    throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
                }

                return data.resolvePhabricatorDiff.oid
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

interface PropsWithInfo {
    info: ConduitDiffDetails
    repoName: string
    filePath: string
    differentialID: number
    diffID: number
    leftDiffID?: number | undefined
    isBase: boolean
    useDiffForBase: boolean
    useBaseForDiff: boolean
}

function getPropsWithInfo(props: ResolveDiffOpt): Promise<PropsWithInfo> {
    return new Promise((resolve, reject) => {
        getDiffDetailsFromConduit(props.diffID, props.differentialID)
            .then(info => ({
                ...props,
                info,
            }))
            .then(propsWithInfo => {
                if (propsWithInfo.isBase || !propsWithInfo.leftDiffID) {
                    // no need to update propsWithInfo
                } else if (
                    hasThisFileChanged(propsWithInfo.filePath, propsWithInfo.info.changes) ||
                    propsWithInfo.isBase ||
                    !propsWithInfo.leftDiffID
                ) {
                    // no need to update propsWithInfo
                } else {
                    getDiffDetailsFromConduit(propsWithInfo.leftDiffID, propsWithInfo.differentialID)
                        .then(info =>
                            resolve({
                                ...propsWithInfo,
                                info,
                                diffID: propsWithInfo.leftDiffID!,
                                useBaseForDiff: true,
                            })
                        )
                        .catch(reject)
                }

                resolve(propsWithInfo)
            })
            .catch(() => {
                // TODO: handle error
            })
    })
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

export function resolveDiffRev(props: ResolveDiffOpt): Observable<ResolvedDiff> {
    // TODO: Do a proper refactor and convert all of this function call and it's deps from Promises to Observables.
    return from<ResolvedDiff>(
        new Promise<ResolvedDiff>((resolve, reject) => {
            getPropsWithInfo(props)
                .then(propsWithInfo => {
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
                    if (!stagingDetails || stagingDetails.unconfigured) {
                        // The last diff (final commit) is not found in the staging area, but rather on the description.
                        if (propsWithInfo.isBase && !propsWithInfo.useDiffForBase) {
                            resolve({ commitID: propsWithInfo.info.sourceControlBaseRevision })
                            return
                        }
                        getRawDiffFromConduit(propsWithInfo.diffID)
                            .then(patch =>
                                resolveStagingRev({ ...conduitProps, patch }).subscribe(commitID => {
                                    if (commitID) {
                                        resolve({ commitID })
                                    }
                                })
                            )
                            .catch(() => reject(new Error('unable to fetch raw diff')))
                        return
                    }

                    if (!stagingDetails.unconfigured) {
                        // Ensure the staging repo exists before resolving. Otherwise create the patch.
                        resolveRepo({ repoName: stagingDetails.repoName }).subscribe(
                            () =>
                                resolve({
                                    commitID: stagingDetails.ref.commit,
                                    stagingRepoName: stagingDetails.repoName,
                                }),
                            error => {
                                getRawDiffFromConduit(propsWithInfo.diffID)
                                    .then(patch =>
                                        resolveStagingRev({ ...conduitProps, patch }).subscribe(commitID => {
                                            if (commitID) {
                                                resolve({ commitID })
                                                return
                                            }
                                            reject(new Error('unable to resolve staging object'))
                                        })
                                    )
                                    .catch(() => reject(new Error('unable to fetch raw diff')))
                            }
                        )
                        return
                    }

                    if (!propsWithInfo.isBase) {
                        for (const cmit of Object.keys(propsWithInfo.info.properties['local:commits'])) {
                            return resolve({ commitID: cmit })
                        }
                    }
                    // last ditch effort to search conduit API for commit ID
                    try {
                        searchForCommitID(propsWithInfo)
                            .then(commitID => resolve({ commitID }))
                            .catch(reject)
                    } catch (e) {
                        // ignore
                    }
                    reject(new Error('did not find commitID'))
                })
                .catch(reject)
        })
    )
}
