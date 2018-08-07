import * as React from 'react'
import { render, unmountComponentAtNode } from 'react-dom'
import { zip } from 'rxjs'
import {
    ChangeState,
    DifferentialState,
    DiffusionState,
    findElementWithOffset,
    getTargetLineAndOffset,
    PhabricatorMode,
    RevisionState,
} from '.'
import { BlobAnnotator } from '../components/BlobAnnotator'
import { fetchBlobContentLines, resolveRepo, resolveRev } from '../repo/backend'
import { getTableDataCell } from '../repo/tooltips'
import {
    ConduitDiffChange,
    ConduitDiffDetails,
    ConduitRef,
    getDiffDetailsFromConduit,
    getRawDiffFromConduit,
    resolveStagingRev,
    searchForCommitID,
} from './backend'
import { StagingAreaInformation } from './components/StagingAreaInformation'
import {
    getCodeCellsForAnnotation,
    getCodeCellsForDifferentialAnnotations,
    getContainerForBlobAnnotation,
    getFilepathFromFile,
    getNodeToConvert,
    getPhabricatorState,
    normalizeRepoPath,
    rowIsNotCode,
    tryGetBlobElement,
} from './util'

const filterFunc = (el: HTMLElement) => !rowIsNotCode(el)

const findTokenCell = (td: HTMLElement, target: HTMLElement) => {
    let curr = target
    while (curr.parentElement && curr.parentElement !== td) {
        curr = curr.parentElement
    }
    return curr
}

/**
 * injectPhabricatorBlobAnnotators finds file blocks on the dom that sould be annotated, and adds blob annotators to them.
 */
export function injectPhabricatorBlobAnnotators(): Promise<void> {
    return getPhabricatorState(window.location).then(state => {
        if (!state) {
            return
        }
        switch (state.mode) {
            case PhabricatorMode.Diffusion:
                return injectDiffusion(state as DiffusionState)

            case PhabricatorMode.Differential:
            case PhabricatorMode.Revision:
            case PhabricatorMode.Change:
                return injectChangeset(state)
        }
    })
}

function createBlobAnnotatorMount(
    fileContainer: HTMLElement,
    actionLinks: Element,
    isBase: boolean
): HTMLElement | null {
    const className = 'sourcegraph-app-annotator' + (isBase ? '-base' : '')
    const existingMount = fileContainer.querySelector('.' + className)
    if (existingMount) {
        // Make this function idempotent; no need to create a mount twice.
        return existingMount as HTMLElement
    }

    const mountEl = document.createElement('div')
    mountEl.style.display = 'inline-block'
    mountEl.className = className

    actionLinks.appendChild(mountEl)
    return mountEl
}

interface ResolveDiffOpt {
    repoPath: string
    filePath: string
    differentialID: number
    diffID: number
    leftDiffID?: number
    isBase: boolean
    useDiffForBase: boolean // indicates whether the base should use the diff commit
    useBaseForDiff: boolean // indicates whether the diff should use the base commit
}

interface ResolvedDiff {
    commitID: string
    stagingRepoPath?: string
}

function monitorFileContainers(
    fileClassName: string,
    tableContainerClassName: string,
    handler: (file: HTMLElement, table: HTMLTableElement, force: boolean) => void
): void {
    const files = document.getElementsByClassName(fileClassName) as HTMLCollectionOf<HTMLElement>

    for (const file of Array.from(files)) {
        const container = file.querySelector(`.${tableContainerClassName}`)

        if (!container) {
            continue
        }

        const observer = new MutationObserver(
            (records: MutationRecord[]): void => {
                for (const rec of records) {
                    for (const n of rec.addedNodes) {
                        const maybeTable = n as HTMLElement
                        if (maybeTable.tagName === 'TABLE') {
                            handler(file, maybeTable as HTMLTableElement, true)
                        }
                    }
                }
            }
        )

        observer.observe(container, { childList: true })

        const table = container.querySelector('table')
        if (table) {
            handler(file, table, false)
        }
    }
}

function injectDiffusion(state: DiffusionState): void {
    const { file, diffusionButtonProps } = getContainerForBlobAnnotation()
    if (!file) {
        console.warn('Unable to find differential blob.')
        return
    }
    const blob = tryGetBlobElement(file)
    if (!blob) {
        console.warn('Unable to find blob element for file', file)
        return
    }
    if (file.className.includes('sg-blob-annotated')) {
        // make this function idempotent
        return
    }
    file.className = `${file.className} sg-blob-annotated`

    const getTableElement = () => file.querySelector('table')!

    const getCodeCells = () => {
        const table = getTableElement()
        if (!table) {
            console.warn('Unable to find table element for file', file)
            return []
        }
        return getCodeCellsForAnnotation(table)
    }
    fetchBlobContentLines(state).subscribe(blobLines => {
        if (blobLines.length === 0) {
            console.warn('Unable to inject blob annotator. No blob lines.')
            return
        }
        const actionLinks =
            file.parentElement!.querySelector('.diffusion-action-bar .phui-right-view') ||
            file.querySelector('.phui-header-action-links')
        if (!actionLinks) {
            console.warn('Unable to find actionLinks', file)
            return
        }
        // Set the parent overflow to visible so that our tooltips show correctly.
        const shellHeader = actionLinks.closest('.phui-header-shell') as HTMLElement
        if (shellHeader) {
            shellHeader.style.overflow = 'visible'
        }
        const mount = createBlobAnnotatorMount(file, actionLinks, true)
        if (!mount) {
            console.warn('Unable to find BlobAnnotatorMount', file)
        }
        render(
            <BlobAnnotator
                getTableElement={getTableElement}
                getCodeCells={getCodeCells}
                getTargetLineAndOffset={getTargetLineAndOffset(blobLines)}
                findElementWithOffset={findElementWithOffset(blobLines)}
                findTokenCell={findTokenCell}
                filterTarget={filterFunc}
                getNodeToConvert={getNodeToConvert}
                fileElement={file}
                repoPath={state.repoPath}
                commitID={state.commitID}
                filePath={state.filePath}
                isPullRequest={false}
                isSplitDiff={false}
                isCommit={false}
                isBase={true}
                buttonProps={diffusionButtonProps}
            />,
            mount
        )
    })
}

function injectChangeset(state: DifferentialState | RevisionState | ChangeState): void {
    monitorFileContainers(
        'differential-changeset',
        'changeset-view-content',
        (file: HTMLElement, table: HTMLTableElement, force: boolean) => {
            if (!force && file.classList.contains('sg-blob-annotated')) {
                // make this function idempotent
                return
            }

            file.classList.add('sg-blob-annotated')

            const actionLinks = file.querySelector('.differential-changeset-buttons') as HTMLElement
            if (!actionLinks) {
                console.warn('Unable to find actionLinks', file)
                return
            }
            actionLinks.style.display = 'inline-flex'

            const mountBase = createBlobAnnotatorMount(file, actionLinks, true)
            if (!mountBase) {
                return
            }
            const mountHead = createBlobAnnotatorMount(file, actionLinks, false)
            if (!mountHead) {
                return
            }
            // TODO(isaac): Find a better way to patch the components.
            // MonitoredBlobAnnotator was not sufficient.
            unmountComponentAtNode(mountBase)
            unmountComponentAtNode(mountHead)

            const { filePath, baseFilePath } = getFilepathFromFile(file)
            const isSplitDiff = table.classList.contains('diff-2up')

            const getTableElement = () => table

            const getCodeCells = (isBase: boolean) => () => {
                const table = getTableElement()
                if (!table) {
                    return []
                }

                return getCodeCellsForDifferentialAnnotations(table, isSplitDiff, isBase)
            }

            const filterTarget = (isBase: boolean) => (target: HTMLElement) => {
                const td = getTableDataCell(target)
                if (!td) {
                    return false
                }
                if (rowIsNotCode(td)) {
                    return false
                }
                if (isSplitDiff) {
                    let curr = td as HTMLElement
                    while (curr.tagName !== 'TH' && !!curr.previousElementSibling) {
                        curr = curr.previousElementSibling as HTMLElement
                    }

                    // Base's line number cell will have no more siblings
                    // while the head's line number cell will.
                    if (isBase) {
                        return !curr.previousSibling
                    }

                    return !!curr.previousSibling
                }
                if (isBase) {
                    return td.classList.contains('left')
                }
                return !td.classList.contains('left')
            }
            const differentialClassname = actionLinks.firstElementChild
                ? `${actionLinks.firstElementChild!.className} msl`
                : 'button grey has-icon msl'
            const differentialButtonProps = {
                className: differentialClassname,
                iconStyle: { marginTop: '-1px', paddingRight: '4px', fontSize: '18px', height: '.8em', width: '.8em' },
                style: {},
            }

            switch (state.mode) {
                case PhabricatorMode.Differential: {
                    const {
                        baseRepoPath,
                        headRepoPath,
                        differentialID,
                        diffID,
                        leftDiffID,
                    } = state as DifferentialState
                    const resolveBaseRevOpt = {
                        repoPath: baseRepoPath,
                        differentialID,
                        diffID: (leftDiffID || diffID)!,
                        leftDiffID,
                        useDiffForBase: Boolean(leftDiffID), // if ?vs and base is not `on` i.e. the initial commit)
                        useBaseForDiff: false,
                        filePath: baseFilePath || filePath,
                        isBase: true,
                        isSplitDiff,
                        filterTarget: filterTarget(true),
                    }
                    const resolveHeadRevOpt = {
                        repoPath: headRepoPath,
                        differentialID,
                        diffID: diffID!,
                        leftDiffID,
                        useDiffForBase: false,
                        useBaseForDiff: false,
                        filePath,
                        isBase: false,
                        isSplitDiff,
                        filterTarget: filterTarget(false),
                    }

                    Promise.all([resolveDiff(resolveBaseRevOpt), resolveDiff(resolveHeadRevOpt)])
                        .then(([baseRev, headRev]) => {
                            const actualBaseRepoPath = baseRev.stagingRepoPath || baseRepoPath
                            const actualHeadRepoPath = headRev.stagingRepoPath || headRepoPath
                            fetchBlobContentLines({
                                repoPath: actualBaseRepoPath,
                                commitID: baseRev.commitID,
                                filePath: baseFilePath || filePath,
                            }).subscribe(baseFile => {
                                if (baseFile.length > 0) {
                                    render(
                                        <BlobAnnotator
                                            {...resolveBaseRevOpt}
                                            repoPath={actualBaseRepoPath}
                                            commitID={baseRev.commitID}
                                            getTargetLineAndOffset={getTargetLineAndOffset(baseFile)}
                                            findElementWithOffset={findElementWithOffset(baseFile)}
                                            findTokenCell={findTokenCell}
                                            getNodeToConvert={getNodeToConvert}
                                            fileElement={file}
                                            isPullRequest={true}
                                            isCommit={false}
                                            buttonProps={differentialButtonProps}
                                            getTableElement={getTableElement}
                                            getCodeCells={getCodeCells(true)}
                                        />,
                                        mountBase
                                    )
                                }
                            })

                            fetchBlobContentLines({
                                repoPath: actualHeadRepoPath,
                                commitID: headRev.commitID,
                                filePath,
                            }).subscribe(headFile => {
                                if (headFile.length > 0) {
                                    render(
                                        <BlobAnnotator
                                            {...resolveHeadRevOpt}
                                            repoPath={actualHeadRepoPath}
                                            commitID={headRev.commitID}
                                            getTargetLineAndOffset={getTargetLineAndOffset(headFile)}
                                            findElementWithOffset={findElementWithOffset(headFile)}
                                            findTokenCell={findTokenCell}
                                            getNodeToConvert={getNodeToConvert}
                                            fileElement={file}
                                            isPullRequest={true}
                                            isCommit={false}
                                            buttonProps={differentialButtonProps}
                                            getTableElement={getTableElement}
                                            getCodeCells={getCodeCells(false)}
                                        />,
                                        mountHead
                                    )
                                }
                            })
                        })
                        .catch(() => {
                            render(<StagingAreaInformation {...differentialButtonProps} />, mountBase)
                        })
                    break // end inner switch
                }

                case PhabricatorMode.Revision: {
                    const { repoPath, baseCommitID, headCommitID } = state as RevisionState
                    zip(
                        fetchBlobContentLines({
                            repoPath,
                            commitID: baseCommitID,
                            filePath,
                        }),
                        fetchBlobContentLines({
                            repoPath,
                            commitID: headCommitID,
                            filePath,
                        })
                    ).subscribe(([baseFile, headFile]) => {
                        if (baseFile.length > 0) {
                            render(
                                <BlobAnnotator
                                    getTargetLineAndOffset={getTargetLineAndOffset(baseFile)}
                                    findElementWithOffset={findElementWithOffset(baseFile)}
                                    findTokenCell={findTokenCell}
                                    getNodeToConvert={getNodeToConvert}
                                    fileElement={file}
                                    repoPath={repoPath}
                                    commitID={baseCommitID}
                                    filePath={filePath}
                                    isPullRequest={true}
                                    isCommit={true}
                                    isBase={true}
                                    isSplitDiff={isSplitDiff}
                                    buttonProps={differentialButtonProps}
                                    getTableElement={getTableElement}
                                    getCodeCells={getCodeCells(true)}
                                    filterTarget={filterTarget(true)}
                                />,
                                mountBase
                            )
                        }
                        if (headFile.length > 0) {
                            render(
                                <BlobAnnotator
                                    getTargetLineAndOffset={getTargetLineAndOffset(headFile)}
                                    findElementWithOffset={findElementWithOffset(headFile)}
                                    findTokenCell={findTokenCell}
                                    getNodeToConvert={getNodeToConvert}
                                    fileElement={file}
                                    repoPath={repoPath}
                                    commitID={headCommitID}
                                    filePath={filePath}
                                    isPullRequest={false}
                                    isCommit={true}
                                    isBase={false}
                                    isSplitDiff={isSplitDiff}
                                    buttonProps={differentialButtonProps}
                                    getTableElement={getTableElement}
                                    getCodeCells={getCodeCells(false)}
                                    filterTarget={filterTarget(false)}
                                />,
                                mountHead
                            )
                        }
                    })
                    break // end inner switch
                }

                case PhabricatorMode.Change: {
                    const { repoPath, commitID } = state as ChangeState
                    resolveRev({ repoPath, rev: commitID + '~1' }).subscribe(baseRev => {
                        Promise.all([
                            fetchBlobContentLines({ repoPath, commitID: baseRev, filePath }).toPromise(),
                            fetchBlobContentLines({ repoPath, commitID, filePath }).toPromise(),
                        ])
                            .then(([baseFile, headFile]) => {
                                if (baseFile.length > 0) {
                                    render(
                                        <BlobAnnotator
                                            getTargetLineAndOffset={getTargetLineAndOffset(baseFile)}
                                            findElementWithOffset={findElementWithOffset(baseFile)}
                                            findTokenCell={findTokenCell}
                                            getNodeToConvert={getNodeToConvert}
                                            fileElement={file}
                                            repoPath={repoPath}
                                            commitID={baseRev}
                                            filePath={filePath}
                                            isPullRequest={true}
                                            isCommit={true}
                                            isBase={true}
                                            isSplitDiff={isSplitDiff}
                                            buttonProps={differentialButtonProps}
                                            getTableElement={getTableElement}
                                            getCodeCells={getCodeCells(true)}
                                            filterTarget={filterTarget(true)}
                                        />,
                                        mountBase
                                    )
                                }
                                if (headFile.length > 0) {
                                    render(
                                        <BlobAnnotator
                                            getTargetLineAndOffset={getTargetLineAndOffset(headFile)}
                                            findElementWithOffset={findElementWithOffset(headFile)}
                                            findTokenCell={findTokenCell}
                                            getNodeToConvert={getNodeToConvert}
                                            fileElement={file}
                                            repoPath={repoPath}
                                            commitID={commitID}
                                            filePath={filePath}
                                            isPullRequest={false}
                                            isCommit={true}
                                            isBase={false}
                                            isSplitDiff={isSplitDiff}
                                            buttonProps={differentialButtonProps}
                                            getTableElement={getTableElement}
                                            getCodeCells={getCodeCells(false)}
                                            filterTarget={filterTarget(false)}
                                        />,
                                        mountHead
                                    )
                                }
                            })
                            .catch(() => {
                                // TODO: handle error
                            })
                    })
                    break // end inner switch
                }
            }
        }
    )
}

function hasThisFileChanged(filePath: string, changes: ConduitDiffChange[]): boolean {
    for (const change of changes) {
        if (change.currentPath === filePath) {
            return true
        }
    }
    return false
}

interface PropsWithInfo {
    info: ConduitDiffDetails
    repoPath: string
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
): { repoPath: string; ref: ConduitRef; unconfigured: boolean } | undefined {
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
                    repoPath: normalizeRepoPath(remote),
                    ref,
                    unconfigured: stagingInfo.status === 'repository.unconfigured',
                }
            }
        }
    }
    return undefined
}

function resolveDiff(props: ResolveDiffOpt): Promise<ResolvedDiff> {
    return new Promise((resolve, reject) => {
        getPropsWithInfo(props)
            .then(propsWithInfo => {
                const stagingDetails = getStagingDetails(propsWithInfo)
                const conduitProps = {
                    repoName: propsWithInfo.repoPath,
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
                    resolveRepo({ repoPath: stagingDetails.repoPath }).subscribe(
                        () =>
                            resolve({ commitID: stagingDetails.ref.commit, stagingRepoPath: stagingDetails.repoPath }),
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
}
