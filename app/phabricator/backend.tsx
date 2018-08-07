import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import storage from '../../extension/storage'
import { getContext } from '../backend/context'
import { mutateGraphQL } from '../backend/graphql'
import { isExtension } from '../context'
import { sourcegraphUrl } from '../util/context'
import { memoizeObservable } from '../util/memoize'
import { normalizeRepoPath } from './util'

interface PhabEntity {
    id: string // e.g. "48"
    type: string // e.g. "RHURI"
    phid: string // e.g. "PHID-RHURI-..."
}

export interface ConduitURI extends PhabEntity {
    fields: {
        uri: {
            raw: string // e.g. https://secure.phabricator.com/source/phabricator.git",
            display: string // e.g. https://secure.phabricator.com/source/phabricator.git",
            effective: string // e.g. https://secure.phabricator.com/source/phabricator.git",
            normalized: string // e.g. secure.phabricator.com/source/phabricator",
        }
    }
}

export interface ConduitRepo extends PhabEntity {
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

export interface ConduitRef {
    ref: string
    type: 'base' | 'diff'
    commit: string // a SHA
    remote: {
        uri: string
    }
}

export interface ConduitDiffChange {
    oldPath: string
    currentPath: string
}

export interface ConduitDiffDetails {
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

export function getDiffDetailsFromConduit(diffID: number, differentialID: number): Promise<ConduitDiffDetails> {
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

export function getRawDiffFromConduit(diffID: number): Promise<string> {
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

export function searchForCommitID(props: any): Promise<string> {
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

export function getRepoPHIDForDifferentialID(differentialID: number): Promise<string> {
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
    repoPath: string
    phabricatorURL: string
}

export const createPhabricatorRepo = memoizeObservable(
    (options: CreatePhabricatorRepoOptions): Observable<void> =>
        mutateGraphQL(
            getContext({ repoKey: options.repoPath, blacklist: ['https://sourcegraph.com'] }),
            `mutation addPhabricatorRepo(
            $callsign: String!,
            $repoPath: String!
            $phabricatorURL: String!
        ) {
            addPhabricatorRepo(callsign: $callsign, uri: $repoPath, url: $phabricatorURL) { alwaysNil }
        }`,
            options
        ).pipe(
            map(({ data, errors }) => {
                if (!data || (errors && errors.length > 0)) {
                    throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
                }
            })
        ),
    ({ callsign }) => callsign
)

export interface PhabricatorRepoDetails {
    callsign: string
    repoPath: string
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
                            repoPath: details.repoPath,
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
        const url = window.localStorage.SOURCEGRAPH_URL || window.SOURCEGRAPH_URL
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

export function getRepoDetailsFromRepoPHID(phid: string): Promise<PhabricatorRepoDetails> {
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
                            repoPath: details.repoPath,
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
                                repoPath: mapping.path,
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
            const callsignMappings =
                window.localStorage.PHABRICATOR_CALLSIGN_MAPPINGS || window.PHABRICATOR_CALLSIGN_MAPPINGS
            const details = convertToDetails(repo)
            if (callsignMappings) {
                for (const mapping of JSON.parse(callsignMappings)) {
                    if (mapping.callsign === repo.fields.callsign) {
                        return resolve({
                            callsign: repo.fields.callsign,
                            repoPath: mapping.path,
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
    const repoPath = normalizeRepoPath(rawURI)
    return { callsign: repo.fields.callsign, repoPath }
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

export const resolveStagingRev = memoizeObservable(
    (options: ResolveStagingOptions): Observable<string | null> =>
        mutateGraphQL(
            getContext({ repoKey: options.repoName, blacklist: ['https://sourcegraph.com'] }),
            `mutation ResolveStagingRev(
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
            options
        ).pipe(
            map(({ data, errors }) => {
                if (!(data && data.resolvePhabricatorDiff) || (errors && errors.length > 0)) {
                    throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
                }

                return data.resolvePhabricatorDiff.oid
            })
        ),
    ({ diffID }: ResolveStagingOptions) => diffID.toString()
)
