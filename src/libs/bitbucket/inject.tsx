import * as React from 'react'
import { render } from 'react-dom'
import { lspViaAPIXlang } from '../../shared/backend/lsp'
import { OpenOnSourcegraph } from '../../shared/components/OpenOnSourcegraph'
import { ServerAuthButton } from '../../shared/components/ServerAuthButton'
import { WithResolvedRev } from '../../shared/components/WithResolvedRev'
import { OpenInSourcegraphProps } from '../../shared/repo'
import { BitbucketMount } from './BitbucketMount'
import { ToolbarActions } from './ToolbarActions'
import { BitbucketRepository, BitbucketState, configureBitbucketHandlers, getRevisionState } from './utils/util'

const BITBUCKET_MOUNT_ID = 'sourcegraph-bitbucket-mount'
const OPEN_ON_SOURCEGRAPH_ID = 'open-on-sourcegraph'

export function injectBitbucketServer(): void {
    configureBitbucketHandlers()
}

function injectBitbucket(state: BitbucketState): void {
    let mount = document.getElementById(BITBUCKET_MOUNT_ID) as HTMLDivElement
    if (!mount) {
        mount = document.createElement('div') as HTMLDivElement
        mount.id = BITBUCKET_MOUNT_ID
        document.body.appendChild(mount)
    }
    if (!state) {
        return configureBitbucketHandlers()
    }
    injectOpenOnSourcegraphButton(state)
    injectBitbucketMount(state, mount)
}

function injectBitbucketMount(state: BitbucketState, mount: HTMLElement): void {
    if (state.repository) {
        let container: HTMLDivElement | undefined
        if (state.pullRequest) {
            const diffFileContainer = getPullRequestCommitContainer()
            if (diffFileContainer) {
                container = diffFileContainer
            } else {
                const pullRequestContainer = getPullRequestActivityContentContainer()
                if (pullRequestContainer) {
                    container = pullRequestContainer
                }
            }
        } else if (state.filePath || state.commit) {
            const fileContainer = getFileContentContainer()
            if (fileContainer) {
                container = fileContainer
            }
        }
        if (!container) {
            console.error('no container found.')
            return
        }
        const observer = new MutationObserver(() => {
            if (injectToolbarActions(state)) {
                observer.disconnect()
            }
        })
        const mainContainer = document.getElementById('page')
        if (mainContainer) {
            observer.observe(mainContainer, { childList: true, subtree: true, characterData: true, attributes: true })
        }
        // Render the View File buttons.
        render(
            <BitbucketMount bitbucketState={state} container={container} simpleProviderFns={lspViaAPIXlang} />,
            mount
        )
    }
}

document.addEventListener('bitbucketLoaded', (e: CustomEvent) => {
    injectBitbucket(e.detail)
})

function getFileToolbarHeaders(): NodeList | undefined {
    return document.querySelectorAll('.file-toolbar') as NodeList
}

function injectToolbarActions(state: BitbucketState): boolean {
    const headers = getFileToolbarHeaders()
    if (!state.repository || !headers) {
        return false
    }
    for (const header of headers) {
        const headerEl = header as HTMLElement
        let filePath = state.filePath ? state.filePath.components.join('/') : undefined
        if (!filePath) {
            const filePathElement = headerEl.querySelector('.breadcrumbs') as HTMLElement
            if (!filePathElement) {
                continue
            }
            filePath = filePathElement.innerText
        }
        let mountEl = headerEl.querySelector('.sourcegraph-app-annotator') as HTMLElement
        if (!mountEl) {
            mountEl = document.createElement('div') as HTMLDivElement
            mountEl.style.display = 'inline-flex'
            mountEl.style.verticalAlign = 'middle'
            mountEl.style.alignItems = 'center'
            mountEl.style.cssFloat = 'left'
            mountEl.className = 'sourcegraph-app-annotator'
            // check if there are secondary buttons to render into or add our own.
            let secondaryButtonGroup = headerEl.querySelector('.secondary') as HTMLElement
            if (!secondaryButtonGroup) {
                secondaryButtonGroup = document.createElement('div')
                secondaryButtonGroup.className = 'secondary'
                const auiButtons = document.createElement('div')
                auiButtons.className = 'aui-buttons'
                secondaryButtonGroup.appendChild(auiButtons)
                header.appendChild(secondaryButtonGroup)
            }
            // Make sure it's first in the button group.
            secondaryButtonGroup.appendChild(mountEl)
        }
        const repoPath = getRepositoryPath(state.repository)
        const revState = getRevisionState(state)
        let rev = 'HEAD'
        if (revState) {
            rev = revState.headRev
        }

        render(
            <WithResolvedRev
                component={ToolbarActions}
                filePath={filePath}
                repoPath={repoPath}
                rev={rev}
                bitbucketState={state}
            />,
            mountEl as HTMLElement
        )
    }
    return true
}

function getRepositoryPath(repository: BitbucketRepository): string {
    return `${window.location.hostname}/${repository.project.key}/${repository.slug}`
}

function injectOpenOnSourcegraphButton(state: BitbucketState): void {
    if (!state.repository) {
        return
    }
    const mount = createOpenButton()
    let margin = '0 10px'
    let actions: HTMLElement | undefined
    if (state.pullRequest) {
        actions = document.querySelector('.pull-request-actions') as HTMLElement
        if (!actions) {
            return
        }
    } else {
        actions = document.querySelector('.repository-breadcrumbs') as HTMLElement
        if (!actions) {
            return
        }
        const innerHeader = actions.closest('.aui-page-header-inner')
        if (!innerHeader) {
            return
        }
        actions = innerHeader.querySelector('.aui-page-header-actions') as HTMLElement
        if (!actions) {
            actions = document.createElement('div')
            actions.className = 'aui-page-header-actions'
            innerHeader.appendChild(actions)
        }
    }

    if (actions.firstChild) {
        actions.insertBefore(mount, actions.firstChild)
    } else {
        margin = '0'
        actions.appendChild(mount)
    }
    const repoPath = getRepositoryPath(state.repository)
    const revState = getRevisionState(state)
    const props: OpenInSourcegraphProps = {
        repoPath,
        rev: revState ? revState.headRev : 'HEAD',
    }
    if (revState && revState.baseRev) {
        props.commit = {
            baseRev: revState.baseRev,
            headRev: revState.headRev,
        }
    }
    render(
        <WithResolvedRev
            component={OpenOnSourcegraph}
            className="aui-button"
            label={props.commit ? 'View Pull Request' : 'View Repository'}
            openProps={props}
            rev={props.rev}
            repoPath={repoPath}
            requireAuthComponent={ServerAuthButton}
            style={{ margin }}
        />,
        mount
    )
}

function createOpenButton(): HTMLElement {
    let container = document.getElementById(OPEN_ON_SOURCEGRAPH_ID)
    if (container) {
        container.remove()
    }
    container = document.createElement('span')
    container.id = OPEN_ON_SOURCEGRAPH_ID
    return container
}

function getPullRequestCommitContainer(): HTMLDivElement | undefined {
    return document.getElementById('commit-file-content') as HTMLDivElement
}

function getPullRequestActivityContentContainer(): HTMLDivElement | undefined {
    return document.querySelector('.pull-request-activity-content') as HTMLDivElement
}

/**
 * getFileContentContainer returns the container for a file that holds line content.
 */
function getFileContentContainer(): HTMLDivElement | undefined {
    return (
        (document.getElementById('file-content') as HTMLDivElement) ||
        (document.getElementById('commit-file-content') as HTMLDivElement)
    )
}
