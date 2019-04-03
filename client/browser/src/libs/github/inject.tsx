import * as React from 'react'
import { render } from 'react-dom'
import { Alerts } from '../../shared/components/Alerts'
import { SymbolsDropdownContainer } from '../../shared/components/SymbolsDropdownContainer'
import { inlineSymbolSearchEnabled } from '../../shared/util/context'
import { querySelectorOrSelf } from '../../shared/util/dom'
import { MountGetter } from '../code_intelligence'
import { getFileContainers, parseURL } from './util'

export async function injectGitHubApplication(marker: HTMLElement): Promise<void> {
    document.body.appendChild(marker)
    await inject()
}

async function inject(): Promise<void> {
    injectServerBanner()

    injectInlineSearch()
}

function injectServerBanner(): void {
    if (window.localStorage['server-banner-enabled'] !== 'true') {
        return
    }

    const { isPullRequest, repoName } = parseURL()
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
    render(<Alerts repoName={repoName} />, mount)
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

        const existingMount = parentDiv.querySelector<HTMLDivElement>(`.${className}`)
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

export const createOpenOnSourcegraphIfNotExists: MountGetter = (container: HTMLElement): HTMLElement | null => {
    const pageheadActions = querySelectorOrSelf(container, '.pagehead-actions')
    // If ran on page that isn't under a repository namespace.
    if (!pageheadActions || pageheadActions.children.length === 0) {
        return null
    }
    // Check for existing
    let mount = pageheadActions.querySelector<HTMLElement>('#' + OPEN_ON_SOURCEGRAPH_ID)
    if (mount) {
        return mount
    }
    // Create new
    mount = document.createElement('li')
    mount.id = OPEN_ON_SOURCEGRAPH_ID
    pageheadActions.insertAdjacentElement('afterbegin', mount)
    return mount
}
