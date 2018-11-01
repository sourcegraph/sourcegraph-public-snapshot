import {
    createHoverifier,
    DOMFunctions,
    findPositionsFromEvents,
    Hoverifier,
    HoverOverlay,
    HoverState,
    LinkComponent,
} from '@sourcegraph/codeintellify'
import { propertyIsDefined } from '@sourcegraph/codeintellify/lib/helpers'
import { HoverMerged } from '@sourcegraph/codeintellify/lib/types'
import {
    Controller as ClientController,
    createController,
} from '@sourcegraph/extensions-client-common/lib/client/controller'
import { Controller } from '@sourcegraph/extensions-client-common/lib/controller'
import { isErrorLike } from '@sourcegraph/extensions-client-common/lib/errors'
import {
    ConfigurationCascadeOrError,
    ConfiguredSubject,
    Settings,
} from '@sourcegraph/extensions-client-common/lib/settings'
import { ConfigurationSubject } from '@sourcegraph/extensions-client-common/lib/settings'
import { ConfigurationCascade } from '@sourcegraph/extensions-client-common/lib/settings'
import * as H from 'history'

import mermaid from 'mermaid'
import * as React from 'react'
import { render, unmountComponentAtNode } from 'react-dom'
import { combineLatest, forkJoin, from, of, Subject } from 'rxjs'
import { filter, map, take, withLatestFrom } from 'rxjs/operators'
import { Disposable } from 'vscode-languageserver'
import { GitHubBlobUrl } from '.'
import { applyDecoration, createMessageTransports } from '../../shared/backend/extensions'
import { createExtensionsContextController } from '../../shared/backend/extensions'
import {
    createJumpURLFetcher,
    createLSPFromExtensions,
    JumpURLLocation,
    lspViaAPIXlang,
    SimpleProviderFns,
    toTextDocumentIdentifier,
} from '../../shared/backend/lsp'
import { Alerts } from '../../shared/components/Alerts'
import { ConfigureSourcegraphButton } from '../../shared/components/ConfigureSourcegraphButton'
import { ContextualSourcegraphButton } from '../../shared/components/ContextualSourcegraphButton'
import { CodeViewToolbar } from '../../shared/components/LegacyCodeViewToolbar'
import { ServerAuthButton } from '../../shared/components/ServerAuthButton'
import { SymbolsDropdownContainer } from '../../shared/components/SymbolsDropdownContainer'
import { WithResolvedRev } from '../../shared/components/WithResolvedRev'
import { AbsoluteRepoFile, DiffResolvedRevSpec } from '../../shared/repo'
import { resolveRev, retryWhenCloneInProgressError } from '../../shared/repo/backend'
import { hideTooltip } from '../../shared/repo/tooltips'
import {
    eventLogger,
    getModeFromPath,
    inlineSymbolSearchEnabled,
    renderMermaidGraphsEnabled,
    sourcegraphUrl,
    useExtensions,
} from '../../shared/util/context'
import { featureFlags } from '../../shared/util/featureFlags'
import { diffDomFunctions, searchCodeSnippetDOMFunctions, singleFileDOMFunctions } from './dom_functions'
import { injectExtensionsGlobalComponents } from './extensions'
import {
    createBlobAnnotatorMount,
    getCodeCommentContainers,
    getDeltaFileName,
    getDiffRepoRev,
    getDiffResolvedRev,
    getFileContainers,
    getGitHubState,
    getRepoCodeSearchContainers,
    parseURL,
} from './util'

const buttonProps = {
    className: 'btn btn-sm tooltipped tooltipped-n',
    style: { marginRight: '5px', textDecoration: 'none', color: 'inherit' },
}

const actionsNavItemClassProps = {
    listClass: 'BtnGroup',
    actionItemClass: 'btn btn-sm tooltipped tooltipped-n BtnGroup-item',
}

function refreshModules(): void {
    for (const el of Array.from(document.getElementsByClassName('sourcegraph-app-annotator'))) {
        el.remove()
    }
    for (const el of Array.from(document.getElementsByClassName('sourcegraph-app-annotator-base'))) {
        el.remove()
    }
    for (const el of Array.from(document.querySelectorAll('.sg-annotated'))) {
        el.classList.remove('sg-annotated')
    }
    hideTooltip()
    inject()
}

window.addEventListener('pjax:end', () => {
    refreshModules()
})

export function injectGitHubApplication(marker: HTMLElement): void {
    document.body.appendChild(marker)
    inject()
}

function injectCodeIntelligence(): void {
    const { repoPath, filePath } = parseURL()

    const files = Array.from(getFileContainers())

    // Heuristic to detect if this page is a single code file (the Sourcegraph extension API currently only
    // supports one file at a time).
    const isSingleCodeFile = files.length === 1 && filePath && document.getElementsByClassName('diff-view').length === 0

    let extensionsContextController: Controller<ConfigurationSubject, Settings> | undefined
    let extensionsController: ClientController<ConfigurationSubject, Settings> | undefined
    let simpleProviderFns = lspViaAPIXlang

    if (isSingleCodeFile && useExtensions && filePath) {
        extensionsContextController = createExtensionsContextController(sourcegraphUrl)
        extensionsController = createController(extensionsContextController.context, createMessageTransports)
        simpleProviderFns = createLSPFromExtensions(extensionsController!)

        const constExtensionsContextController = extensionsContextController!
        const constController = extensionsController!

        injectExtensionsGlobalComponents(
            {
                extensionsController: constController,
                extensionsContextController: constExtensionsContextController,
            },
            H.createLocation(window.location)
        )

        resolveRev({ repoPath, rev: parseURL().rev })
            .pipe(retryWhenCloneInProgressError())
            .subscribe(commitID => {
                combineLatest(
                    constExtensionsContextController.viewerConfiguredExtensions,
                    constExtensionsContextController.context.configurationCascade
                ).subscribe(
                    ([configuredExtensions, configurationCascade]) => {
                        const toURIWithPath = (ctx: AbsoluteRepoFile) =>
                            `git://${ctx.repoPath}?${ctx.commitID}#${ctx.filePath}`

                        const fileElement = document.querySelector('tbody')
                        const gitHubCurrentFileContent = fileElement ? fileElement.innerText : ''

                        // This is rather specific to extensions-client-common
                        // and could be moved to that package in the future.
                        const logThenDropConfigurationErrors = (
                            cascadeOrError: ConfigurationCascadeOrError<ConfigurationSubject, Settings>
                        ): ConfigurationCascade<ConfigurationSubject, Settings> => {
                            const EMPTY_CASCADE: ConfigurationCascade<ConfigurationSubject, Settings> = {
                                subjects: [],
                                merged: {},
                            }
                            if (!cascadeOrError.subjects) {
                                console.error('invalid configuration: no configuration subjects available')
                                return EMPTY_CASCADE
                            }
                            if (!cascadeOrError.merged) {
                                console.error('invalid configuration: no merged configuration available')
                                return EMPTY_CASCADE
                            }
                            if (isErrorLike(cascadeOrError.subjects)) {
                                console.error(
                                    `invalid configuration: error in configuration subjects: ${
                                        cascadeOrError.subjects.message
                                    }`
                                )
                                return EMPTY_CASCADE
                            }
                            if (isErrorLike(cascadeOrError.merged)) {
                                console.error(
                                    `invalid configuration: error in merged configuration: ${
                                        cascadeOrError.merged.message
                                    }`
                                )
                                return EMPTY_CASCADE
                            }
                            return {
                                subjects: cascadeOrError.subjects.filter(
                                    (subject): subject is ConfiguredSubject<ConfigurationSubject, Settings> => {
                                        if (!subject) {
                                            console.error('invalid configuration: no configuration subjects available')
                                            return false
                                        }
                                        if (isErrorLike(subject)) {
                                            console.error(
                                                `invalid configuration: error in configuration subjects: ${
                                                    subject.message
                                                }`
                                            )
                                            return false
                                        }
                                        return true
                                    }
                                ),
                                merged: cascadeOrError.merged,
                            }
                        }

                        from(constController.environment)
                            .pipe(take(1))
                            .subscribe(previous => {
                                constController.setEnvironment({
                                    visibleTextDocuments: [
                                        {
                                            uri: toURIWithPath({ repoPath, commitID, filePath }),
                                            languageId: getModeFromPath(filePath) || 'could not determine mode',
                                            text: gitHubCurrentFileContent,
                                        },
                                    ],
                                    extensions: configuredExtensions,
                                    configuration: logThenDropConfigurationErrors(configurationCascade),
                                    context: previous.context,
                                })
                            })

                        let oldDecorations: Disposable[] = []

                        constController.registries.textDocumentDecoration
                            .getDecorations(
                                toTextDocumentIdentifier({
                                    commitID,
                                    filePath,
                                    repoPath,
                                })
                            )
                            .subscribe(decorations => {
                                for (const old of oldDecorations) {
                                    old.dispose()
                                }
                                oldDecorations = []
                                for (const decoration of decorations || []) {
                                    try {
                                        oldDecorations.push(
                                            applyDecoration({
                                                fileElement: files[0],
                                                decoration,
                                            })
                                        )
                                    } catch (e) {
                                        console.warn(e)
                                    }
                                }
                            })
                    },
                    err => {
                        console.error('Error fetching viewer configured extensions via GraphQL: %O', err)
                    }
                )
            })
    }

    const hoverifier = createCodeIntelligenceContainer({ repoPath, simpleProviderFns })

    injectBlobAnnotators(hoverifier, files, lspViaAPIXlang, extensionsContextController, extensionsController)

    injectCodeSnippetAnnotator(hoverifier, getCodeCommentContainers(), '.border.rounded-1.my-2', singleFileDOMFunctions)
    injectCodeSnippetAnnotator(
        hoverifier,
        getRepoCodeSearchContainers(),
        '.d-inline-block',
        searchCodeSnippetDOMFunctions
    )
}

function inject(): void {
    featureFlags
        .isEnabled('newInject')
        .then(isEnabled => {
            if (!isEnabled) {
                injectCodeIntelligence()
            }
        })
        .catch(err => console.error('could not get feature flag', err))

    injectServerBanner()
    injectOpenOnSourcegraphButton()

    injectMermaid()

    injectInlineSearch()
}

/**
 * injectCodeSnippetAnnotator annotates the given containers and adds a view file button.
 * @param containers The blob containers that holds the code snippet to be annotated.
 * @param selector The selector of the element to append a "View File" button.
 */
function injectCodeSnippetAnnotator(
    hoverifier: Hoverifier,
    containers: HTMLCollectionOf<HTMLElement>,
    selector: string,
    dom: DOMFunctions
): void {
    for (const file of containers) {
        const filePathContainer = file.querySelector(selector)
        if (!filePathContainer) {
            continue
        }

        const anchors = file.getElementsByTagName('a')
        let gitHubState: GitHubBlobUrl | undefined
        for (const anchor of Array.from(anchors)) {
            const anchorState = getGitHubState(anchor.href) as GitHubBlobUrl
            if (anchorState) {
                gitHubState = anchorState
                break
            }
        }

        if (!gitHubState || !gitHubState.owner || !gitHubState.repoName || !gitHubState.rev || !gitHubState.filePath) {
            continue
        }

        const { rev, filePath, repoName, owner } = gitHubState!
        const repoPath = `${window.location.host}/${owner}/${repoName}`

        const mount = document.createElement('div')
        // mount.style.display = 'none'
        mount.className = 'sourcegraph-app-annotator'
        filePathContainer.appendChild(mount)

        const location = H.createLocation(window.location)

        const renderCodeView = (commitID: string) =>
            render(
                <CodeViewToolbar
                    repoPath={repoPath}
                    filePath={filePath}
                    baseCommitID={commitID}
                    baseRev={commitID}
                    buttonProps={buttonProps}
                    simpleProviderFns={lspViaAPIXlang}
                    actionsNavItemClassProps={actionsNavItemClassProps}
                    location={location}
                />,
                mount
            )

        resolveRev({ repoPath, rev })
            .pipe(retryWhenCloneInProgressError())
            .subscribe(
                commitID => {
                    hoverifier.hoverify({
                        dom,
                        positionEvents: of(file).pipe(findPositionsFromEvents(dom)),
                        resolveContext: () => ({
                            repoPath,
                            filePath: filePath!,
                            rev: rev || commitID,
                            commitID,
                        }),
                    })

                    renderCodeView(commitID)
                },
                err => console.error(repoPath, rev, err)
            )
    }
}

function injectServerBanner(): void {
    if (window.localStorage['server-banner-enabled'] !== 'true') {
        return
    }

    const { isPullRequest, repoPath } = parseURL()
    if (!isPullRequest) {
        return
    }
    // Check which files were modified.
    const files = getFileContainers()
    if (!files.length) {
        return
    }

    let mount = document.getElementById('server-alert-mount')
    if (!mount) {
        mount = document.createElement('div')
        mount.id = 'server-alert-mount'
        const container = document.getElementById('partial-discussion-header')
        if (!container) {
            return
        }
        container.appendChild(mount)
    }
    render(<Alerts repoPath={repoPath} />, mount)
}

function createCodeIntelligenceContainer(options: {
    repoPath: string
    simpleProviderFns: SimpleProviderFns
}): Hoverifier {
    const overlayMountID = 'sg-tooltip-mount'

    let overlayMount = document.getElementById(overlayMountID)
    if (!overlayMount) {
        overlayMount = document.createElement('div')
        overlayMount.id = overlayMountID
        overlayMount.style.height = '0px'
        document.body.appendChild(overlayMount)
    } else {
        // "Force reload" code intelligence.
        unmountComponentAtNode(overlayMount)
        for (const highlighted of document.querySelectorAll('.selection-highlight')) {
            highlighted.classList.remove('selection-highlight')
        }
    }

    /** Emits when the go to definition button was clicked */
    const goToDefinitionClicks = new Subject<MouseEvent>()
    const nextGoToDefinitionClick = (event: MouseEvent) => goToDefinitionClicks.next(event)

    /** Emits when the close button was clicked */
    const closeButtonClicks = new Subject<MouseEvent>()
    const nextCloseButtonClick = (event: MouseEvent) => closeButtonClicks.next(event)

    /** Emits whenever the ref callback for the hover element is called */
    const hoverOverlayElements = new Subject<HTMLElement | null>()
    const nextOverlayElement = (element: HTMLElement | null) => hoverOverlayElements.next(element)

    const relativeElement = document.body

    const fetchJumpURL = createJumpURLFetcher(options.simpleProviderFns.fetchDefinition, (def: JumpURLLocation) => {
        const rev = def.rev
        // If we're provided options, we can make the j2d URL more specific.
        if (options) {
            const sameRepo = options.repoPath === def.repoPath
            // Stay on same page in PR if possible.
            if (sameRepo && def.part) {
                const containers = getFileContainers()
                for (const container of containers) {
                    const header = container.querySelector('.file-header') as HTMLElement
                    const anchorPath = header.dataset.path
                    if (anchorPath === def.filePath) {
                        const anchorUrl = header.dataset.anchor
                        const url = `${window.location.origin}${window.location.pathname}#${anchorUrl}${
                            def.part === 'base' ? 'L' : 'R'
                        }${def.position.line}`

                        return url
                    }
                }
            }
        }

        return `https://${def.repoPath}/blob/${rev}/${def.filePath}#L${def.position.line}${
            def.position.character ? ':' + def.position.character : ''
        }`
    })

    const containerComponentUpdates = new Subject<void>()

    const hoverifier = createHoverifier({
        closeButtonClicks,
        goToDefinitionClicks,
        hoverOverlayElements,
        hoverOverlayRerenders: containerComponentUpdates.pipe(
            withLatestFrom(hoverOverlayElements),
            map(([, hoverOverlayElement]) => ({ hoverOverlayElement, relativeElement })),
            filter(propertyIsDefined('hoverOverlayElement'))
        ),
        pushHistory: path => {
            location.href = path
        },
        fetchHover: ({ line, character, part, ...rest }) =>
            options.simpleProviderFns
                .fetchHover({ ...rest, position: { line, character } })
                .pipe(map(hover => (hover ? (hover as HoverMerged) : hover))),
        fetchJumpURL,
        logTelemetryEvent: () => eventLogger.logCodeIntelligenceEvent(),
    })

    class HoverOverlayContainer extends React.Component<{}, HoverState> {
        constructor(props: {}) {
            super(props)
            this.state = hoverifier.hoverState
            hoverifier.hoverStateUpdates.subscribe(update => this.setState(update))
        }
        public componentDidMount(): void {
            containerComponentUpdates.next()
        }
        public componentDidUpdate(): void {
            containerComponentUpdates.next()
        }
        public render(): JSX.Element | null {
            return this.state.hoverOverlayProps ? (
                <HoverOverlay
                    {...this.state.hoverOverlayProps}
                    linkComponent={LinkComponent}
                    logTelemetryEvent={this.log}
                    hoverRef={nextOverlayElement}
                    onGoToDefinitionClick={nextGoToDefinitionClick}
                    onCloseButtonClick={nextCloseButtonClick}
                />
            ) : null
        }
        private log = () => eventLogger.logCodeIntelligenceEvent()
    }

    render(<HoverOverlayContainer />, overlayMount)

    return hoverifier
}

const LinkComponent: LinkComponent = ({ to, children, ...rest }) => (
    <a href={new URL(to, sourcegraphUrl).href} {...rest}>
        {children}
    </a>
)

function injectBlobAnnotators(
    hoverifier: Hoverifier,
    files: HTMLElement[],
    simpleProviderFns: SimpleProviderFns,
    extensions?: Controller<ConfigurationSubject, Settings>,
    extensionsController?: ClientController<ConfigurationSubject, Settings>
): void {
    const { repoPath, isDelta, filePath, rev } = parseURL()
    if (!filePath && !isDelta) {
        return
    }

    function addBlobAnnotator(file: HTMLElement, hoverifier: Hoverifier, simpleProviderFns: SimpleProviderFns): void {
        const diffLoader = file.querySelector('.js-diff-load-container')
        if (diffLoader) {
            const observer = new MutationObserver(() => {
                const element = diffLoader.querySelector('.diff-table')
                if (element) {
                    addBlobAnnotator(file, hoverifier, simpleProviderFns)
                    observer.disconnect()
                }
            })
            observer.observe(diffLoader, { childList: true })
        }

        if (!isDelta) {
            const mount = createBlobAnnotatorMount(file)
            if (!mount) {
                return
            }

            const renderCodeView = (commitID: string) =>
                render(
                    <CodeViewToolbar
                        repoPath={repoPath}
                        filePath={filePath!}
                        baseCommitID={commitID}
                        baseRev={commitID}
                        buttonProps={buttonProps}
                        simpleProviderFns={simpleProviderFns}
                        extensionsController={extensionsController}
                        extensions={extensions}
                        actionsNavItemClassProps={actionsNavItemClassProps}
                        location={H.createLocation(window.location)}
                    />,
                    mount
                )

            resolveRev({ repoPath, rev })
                .pipe(retryWhenCloneInProgressError())
                .subscribe(
                    commitID => {
                        hoverifier.hoverify({
                            dom: singleFileDOMFunctions,
                            positionEvents: of(file).pipe(findPositionsFromEvents(singleFileDOMFunctions)),
                            resolveContext: () => ({
                                repoPath,
                                filePath: filePath!,
                                rev: rev || commitID,
                                commitID,
                            }),
                        })
                        renderCodeView(commitID)
                    },
                    err => console.error(err)
                )

            return
        }

        const { headFilePath, baseFilePath } = getDeltaFileName(file)
        if (!headFilePath) {
            console.error('cannot determine file path')
            return
        }

        let baseCommitID: string
        let headCommitID: string
        const deltaRevs = getDiffResolvedRev()
        if (!deltaRevs) {
            console.error('cannot determine deltaRevs')
            return
        }

        baseCommitID = deltaRevs.baseCommitID
        headCommitID = deltaRevs.headCommitID

        const deltaInfo = getDiffRepoRev()
        if (!deltaInfo) {
            console.error('cannot determine deltaInfo')
            return
        }

        forkJoin(
            resolveRev({ repoPath, rev: baseCommitID }).pipe(retryWhenCloneInProgressError()),
            resolveRev({ repoPath, rev: headCommitID }).pipe(retryWhenCloneInProgressError())
        )
            .pipe(
                map(
                    ([baseCommitID, headCommitID]): DiffResolvedRevSpec => ({
                        baseCommitID,
                        headCommitID,
                    })
                )
            )
            .subscribe(resolvedRevSpec => {
                const mount = createBlobAnnotatorMount(file, true)
                if (mount) {
                    render(
                        <CodeViewToolbar
                            repoPath={repoPath}
                            filePath={headFilePath}
                            baseCommitID={resolvedRevSpec.baseCommitID}
                            headCommitID={resolvedRevSpec.headCommitID}
                            buttonProps={buttonProps}
                            simpleProviderFns={simpleProviderFns}
                            actionsNavItemClassProps={actionsNavItemClassProps}
                            location={H.createLocation(window.location)}
                        />,
                        mount
                    )
                }

                hoverifier.hoverify({
                    dom: diffDomFunctions,
                    positionEvents: of(file).pipe(findPositionsFromEvents(diffDomFunctions)),
                    resolveContext: ({ part }) => ({
                        repoPath,
                        rev: part === 'base' ? resolvedRevSpec.baseCommitID : resolvedRevSpec.headCommitID,
                        commitID: part === 'base' ? resolvedRevSpec.baseCommitID : resolvedRevSpec.headCommitID,
                        // If a hover happened on the base, it must exist
                        filePath: part === 'base' ? baseFilePath! : headFilePath,
                    }),
                })
            })
    }

    for (const file of files) {
        addBlobAnnotator(file as HTMLElement, hoverifier, simpleProviderFns)
    }
    const mutationObserver = new MutationObserver(mutations => {
        for (const mutation of mutations) {
            const nodes = Array.prototype.slice.call(mutation.addedNodes)
            for (const node of nodes) {
                if (node && node.classList && node.classList.contains('file') && node.classList.contains('js-file')) {
                    const intersectionObserver = new IntersectionObserver(
                        entries => {
                            for (const file of entries) {
                                // File is an IntersectionObserverEntry, which has `isIntersecting` as a prop, but TS
                                // complains that it does not exist.
                                if ((file as any).isIntersecting && !file.target.classList.contains('annotated')) {
                                    file.target.classList.add('annotated')
                                    addBlobAnnotator(file.target as HTMLElement, hoverifier, simpleProviderFns)
                                }
                            }
                        },
                        {
                            rootMargin: '200px',
                            threshold: 0,
                        }
                    )
                    intersectionObserver.observe(node)
                }
            }
        }
    })
    const filebucket = document.getElementById('files')
    if (!filebucket) {
        return
    }

    mutationObserver.observe(filebucket, {
        childList: true,
        subtree: true,
        attributes: false,
        characterData: false,
    })
}

/**
 * Appends an Open on Sourcegraph button to the GitHub DOM.
 * The button is only rendered on a repo homepage after the "find file" button.
 */
function injectOpenOnSourcegraphButton(): void {
    const container = createOpenOnSourcegraphIfNotExists()
    const pageheadActions = document.querySelector('.pagehead-actions')
    if (!pageheadActions || !pageheadActions.children.length) {
        return
    }
    pageheadActions.insertBefore(container, pageheadActions.children[0])
    if (container) {
        const { repoPath, rev } = parseURL()
        if (repoPath) {
            render(
                <WithResolvedRev
                    component={ContextualSourcegraphButton}
                    repoPath={repoPath}
                    rev={rev}
                    defaultBranch={'HEAD'}
                    notFoundComponent={ConfigureSourcegraphButton}
                    requireAuthComponent={ServerAuthButton}
                />,
                container
            )
        }
    }
}

function injectMermaid(): void {
    if (!renderMermaidGraphsEnabled) {
        return
    }

    // The structure looks like:
    //
    //    ...
    //    <pre lang="mermaid">
    //       <code>
    //          graph TD;
    //             A-->B;
    //       </code>
    //    </pre>
    //   ...
    //
    // We want to end up with:
    //
    //    ...
    //    <pre lang="mermaid">
    //       <code>
    //          graph TD;
    //             A-->B;
    //       </code>
    //    </pre>
    //    <svg>
    //       /* SVG FROM MERMAID GOES HERE */
    //    </svg>
    //   ...

    let id = 1

    const renderMermaidCharts = () => {
        const pres = document.querySelectorAll('pre[lang=mermaid]')
        for (const pre of pres) {
            const el = pre as HTMLElement
            if (el.style.display === 'none') {
                // already rendered
                continue
            }
            el.style.display = 'none'
            const chartDefinition = pre.getElementsByTagName('code')[0].textContent || ''
            const chartID = `mermaid_${id++}`
            mermaid.mermaidAPI.render(chartID, chartDefinition, svg => el.insertAdjacentHTML('afterend', svg))
        }
    }

    // Render mermaid charts async and debounce the rendering
    // to minimize impact on page load.
    let timeout: number | undefined
    const handleDomChange = () => {
        clearTimeout(timeout)
        // Need to use window.setTimeout because:
        // https://github.com/DefinitelyTyped/DefinitelyTyped/issues/21310#issuecomment-367919251
        timeout = window.setTimeout(() => renderMermaidCharts(), 100)
    }

    const observer = new MutationObserver(() => handleDomChange())
    observer.observe(document.body, { subtree: true, childList: true })
    handleDomChange()
}

function injectInlineSearch(): void {
    if (!inlineSymbolSearchEnabled) {
        return
    }

    // idempotently create a div to render the autocomplete react component inside of
    function createAutoCompleteContainerMount(textArea: HTMLTextAreaElement): HTMLDivElement | undefined {
        const parentDiv = textArea.parentElement
        if (!parentDiv) {
            return undefined
        }

        const className = 'symbols-autocomplete'

        const existingMount = parentDiv.querySelector(`.${className}`) as HTMLDivElement | null
        if (existingMount) {
            return existingMount
        }

        const mountElement = document.createElement('div')
        mountElement.className = className
        parentDiv.appendChild(mountElement)

        return mountElement
    }

    // lazily attach the symbols dropdown container whenever
    // a text area is focused
    document.addEventListener('focusin', e => {
        if (!e.target) {
            return
        }

        const target = e.target as HTMLElement

        if (target.tagName !== 'TEXTAREA') {
            return
        }

        const textArea = target as HTMLTextAreaElement
        const mountElement = createAutoCompleteContainerMount(textArea)
        if (mountElement) {
            render(<SymbolsDropdownContainer textBoxRef={textArea} />, mountElement)
        }
    })
}

const OPEN_ON_SOURCEGRAPH_ID = 'open-on-sourcegraph'

function createOpenOnSourcegraphIfNotExists(): HTMLElement {
    let container = document.getElementById(OPEN_ON_SOURCEGRAPH_ID)
    if (container) {
        container.remove()
    }

    container = document.createElement('li')
    container.id = OPEN_ON_SOURCEGRAPH_ID
    return container
}
