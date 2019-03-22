import { AdjustmentDirection, DOMFunctions, PositionAdjuster } from '@sourcegraph/codeintellify'
import { of } from 'rxjs'
import { FileSpec, RepoSpec, ResolvedRevSpec, RevSpec } from '../../../../../shared/src/util/url'
import { CodeHost, CodeViewSpecResolver, CodeViewSpecWithOutSelector } from '../code_intelligence'
import { diffDOMFunctions, singleFileDOMFunctions } from './dom_functions'
import { resolveDiffFileInfo, resolveFileInfo } from './file_info'

const createToolbarMount = (codeView: HTMLElement) => {
    const existingMount = codeView.querySelector<HTMLElement>('.sg-toolbar-mount')
    if (existingMount) {
        return existingMount
    }

    const fileActions = codeView.querySelector<HTMLElement>('.file-toolbar .secondary')
    if (!fileActions) {
        throw new Error('Unable to find mount location')
    }

    fileActions.style.display = 'flex'

    const mount = document.createElement('div')
    mount.classList.add('btn-group')
    mount.classList.add('sg-toolbar-mount')
    mount.classList.add('sg-toolbar-mount-bitbucket-server')

    fileActions.insertAdjacentElement('afterbegin', mount)

    return mount
}

/**
 * Sometimes tabs are converted to spaces so we need to adjust. Luckily, there
 * is an attribute `cm-text` that contains the real text.
 */
const createPositionAdjuster = (
    dom: DOMFunctions
): PositionAdjuster<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec> => ({ direction, codeView, position }) => {
    const codeElement = dom.getCodeElementFromLineNumber(codeView, position.line, position.part)
    if (!codeElement) {
        throw new Error('(adjustPosition) could not find code element for line provided')
    }

    let delta = 0
    for (const modifiedTextElem of codeElement.querySelectorAll('[cm-text]')) {
        const actualText = modifiedTextElem.getAttribute('cm-text') || ''
        const adjustedText = modifiedTextElem.textContent || ''

        delta += actualText.length - adjustedText.length
    }

    const modifier = direction === AdjustmentDirection.ActualToCodeView ? -1 : 1

    const newPos = {
        line: position.line,
        character: position.character + modifier * delta,
    }

    return of(newPos)
}

const singleFileCodeView: CodeViewSpecWithOutSelector = {
    getToolbarMount: createToolbarMount,
    dom: singleFileDOMFunctions,
    resolveFileInfo,
    adjustPosition: createPositionAdjuster(singleFileDOMFunctions),
    toolbarButtonProps: {
        className: 'aui-button',
        style: { marginLeft: 10 },
    },
}

const diffCodeView: CodeViewSpecWithOutSelector = {
    getToolbarMount: createToolbarMount,
    dom: diffDOMFunctions,
    resolveFileInfo: resolveDiffFileInfo,
    adjustPosition: createPositionAdjuster(diffDOMFunctions),
    toolbarButtonProps: {
        className: 'aui-button',
        style: { marginLeft: 10 },
    },
}

const resolveCodeViewSpec: CodeViewSpecResolver['resolveCodeViewSpec'] = codeView => {
    const contentView = codeView.querySelector('.content-view')
    if (!contentView) {
        return null
    }

    const isDiff = contentView.classList.contains('diff-view')

    return isDiff ? diffCodeView : singleFileCodeView
}

const codeViewResolver: CodeViewSpecResolver = {
    selector: '.file-content',
    resolveCodeViewSpec,
}

function getCommandPaletteMount(): HTMLElement {
    const headerElem = document.querySelector('.aui-header-primary .aui-nav')
    if (!headerElem) {
        throw new Error('Unable to find command palette mount')
    }

    const commandListClass = 'command-palette-button command-palette-button__bitbucket-server'

    const createCommandList = (): HTMLElement => {
        const commandListElem = document.createElement('li')
        commandListElem.className = commandListClass
        headerElem!.insertAdjacentElement('beforeend', commandListElem)

        return commandListElem
    }

    return document.querySelector<HTMLElement>('.' + commandListClass) || createCommandList()
}

export const bitbucketServerCodeHost: CodeHost = {
    name: 'bitbucket-server',
    check: () =>
        !!document.querySelector('.bitbucket-header-logo') ||
        !!document.querySelector('.aui-header-logo.aui-header-logo-bitbucket'),
    codeViewSpecResolver: codeViewResolver,
    getCommandPaletteMount,
}
