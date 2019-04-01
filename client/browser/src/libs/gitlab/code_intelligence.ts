import { CodeHost, CodeViewSpecResolver, CodeViewSpecWithOutSelector } from '../code_intelligence'
import { diffDOMFunctions, singleFileDOMFunctions } from './dom_functions'
import { getCommandPaletteMount } from './extensions'
import { resolveCommitFileInfo, resolveDiffFileInfo, resolveFileInfo } from './file_info'
import { getPageInfo, GitLabPageKind } from './scrape'

const toolbarButtonProps = {
    className: 'btn btn-default btn-sm',
    style: { marginRight: '5px', textDecoration: 'none', color: 'inherit' },
}

export function checkIsGitlab(): boolean {
    return !!document.head!.querySelector('meta[content="GitLab"]')
}

const adjustOverlayPosition: CodeHost['adjustOverlayPosition'] = ({ top, left }) => {
    const header = document.querySelector('header')

    return {
        top: header ? top + header.getBoundingClientRect().height : 0,
        left,
    }
}

const createToolbarMount = (codeView: HTMLElement) => {
    const existingMount: HTMLElement | null = codeView.querySelector('.sg-toolbar-mount-gitlab')
    if (existingMount) {
        return existingMount
    }

    const fileActions = codeView.querySelector('.file-actions')
    if (!fileActions) {
        throw new Error('Unable to find mount location')
    }

    const mount = document.createElement('div')
    mount.classList.add('btn-group')
    mount.classList.add('sg-toolbar-mount')
    mount.classList.add('sg-toolbar-mount-gitlab')

    fileActions.insertAdjacentElement('afterbegin', mount)

    return mount
}

const singleFileCodeView: CodeViewSpecWithOutSelector = {
    dom: singleFileDOMFunctions,
    isDiff: false,
    getToolbarMount: createToolbarMount,
    resolveFileInfo,
    toolbarButtonProps,
}

const mergeRequestCodeView: CodeViewSpecWithOutSelector = {
    dom: diffDOMFunctions,
    isDiff: true,
    getToolbarMount: createToolbarMount,
    resolveFileInfo: resolveDiffFileInfo,
    toolbarButtonProps,
}

const commitCodeView: CodeViewSpecWithOutSelector = {
    dom: diffDOMFunctions,
    isDiff: true,
    getToolbarMount: createToolbarMount,
    resolveFileInfo: resolveCommitFileInfo,
    toolbarButtonProps,
}

const resolveCodeViewSpec = (codeView: HTMLElement): CodeViewSpecWithOutSelector => {
    const { pageKind } = getPageInfo()

    if (pageKind === GitLabPageKind.File) {
        return singleFileCodeView
    }

    if (pageKind === GitLabPageKind.MergeRequest) {
        return mergeRequestCodeView
    }

    return commitCodeView
}

const codeViewResolver: CodeViewSpecResolver = {
    selector: '.file-holder',
    resolveCodeViewSpec,
}

export const gitlabCodeHost: CodeHost = {
    name: 'gitlab',
    check: checkIsGitlab,
    codeViewSpecResolver: codeViewResolver,
    adjustOverlayPosition,
    getCommandPaletteMount,
    actionNavItemClassProps: {
        actionItemClass: 'btn btn-secondary action-item--gitlab',
        actionItemPressedClass: 'active',
    },
}
