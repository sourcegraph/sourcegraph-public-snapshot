import { decode } from 'js-base64'

import { SearchPatternType } from '../graphql-operations'
import type { PreviewRequest, Request } from '../search/js-to-java-bridge'
import type { Search, Theme } from '../search/types'

import { dark } from './theme-snapshots/dark'
import { light } from './theme-snapshots/light'

/* Set these to connect to a different server */
const instanceURL = 'https://sourcegraph.com/'
const accessToken = null

let isDarkTheme = false

const savedSearchFromLocalStorage = localStorage.getItem('savedSearch')
let savedSearch: Search = savedSearchFromLocalStorage
    ? (JSON.parse(savedSearchFromLocalStorage) as Search)
    : {
          query: 'r:github.com/sourcegraph/sourcegraph jetbrains',
          caseSensitive: false,
          patternType: SearchPatternType.literal,
          selectedSearchContextSpec: 'global',
      }

const webviewOverlay = document.querySelector('#webview-overlay') as HTMLPreElement
const codeDetailsNode = document.querySelector('#code-details') as HTMLPreElement

let previewContent: PreviewRequest['arguments'] | null = null

export function callJava(request: Request): Promise<object> {
    return new Promise((resolve, reject) => {
        const requestAsString = JSON.stringify(request)
        const onSuccessCallback = (responseAsString: string): void => {
            resolve(JSON.parse(responseAsString))
        }
        const onFailureCallback = (errorCode: number, errorMessage: string): void => {
            reject(new Error(`${errorCode} - ${errorMessage}`))
        }
        console.log(`The mock Java backend just received this request: ${requestAsString}`)
        handleRequest(request, onSuccessCallback, onFailureCallback)
    })
}

function handleRequest(
    request: Request,
    onSuccessCallback: (responseAsString: string) => void,
    onFailureCallback: (errorCode: number, errorMessage: string) => void
): void {
    const action = request.action
    switch (action) {
        case 'getConfig': {
            onSuccessCallback(
                JSON.stringify({
                    instanceURL,
                    accessToken,
                    customRequestHeadersAsString: '',
                    pluginVersion: '1.2.3',
                    anonymousUserId: 'test',
                })
            )
            break
        }

        case 'getTheme': {
            const theme: Theme = isDarkTheme ? dark : light
            onSuccessCallback(JSON.stringify(theme))
            break
        }

        case 'indicateSearchError': {
            setOverlay(`Search error: ${request.arguments.errorMessage}`)
            onSuccessCallback('null')
            break
        }

        case 'previewLoading': {
            setOverlay(null)
            codeDetailsNode.innerHTML = 'Loading...'
            onSuccessCallback('null')
            break
        }

        case 'preview': {
            setOverlay(null)
            previewContent = request.arguments

            const start =
                previewContent.absoluteOffsetAndLengths && previewContent.absoluteOffsetAndLengths.length > 0
                    ? previewContent.absoluteOffsetAndLengths[0][0]
                    : 0
            const length =
                previewContent.absoluteOffsetAndLengths && previewContent.absoluteOffsetAndLengths.length > 0
                    ? previewContent.absoluteOffsetAndLengths[0][1]
                    : 0

            let htmlContent: string
            if (previewContent.content === null) {
                htmlContent = 'No preview available'
            } else {
                const decodedContent = decode(previewContent.content || '')
                htmlContent = escapeHTML(decodedContent.slice(0, start))
                htmlContent += `<span id="code-details-highlight">${escapeHTML(
                    decodedContent.slice(start, start + length)
                )}</span>`
                htmlContent += escapeHTML(decodedContent.slice(start + length))
            }

            codeDetailsNode.innerHTML = htmlContent

            document.querySelector('#code-details-highlight')?.scrollIntoView({ block: 'center', inline: 'center' })

            onSuccessCallback('null')
            break
        }

        case 'clearPreview': {
            setOverlay(null)
            codeDetailsNode.textContent = ''
            onSuccessCallback('null')
            break
        }

        case 'open': {
            previewContent = request.arguments
            if (previewContent.fileName) {
                alert(`Now the IDE would open ${previewContent.path} in the editor...`)
            } else {
                window.open(instanceURL + (previewContent.relativeUrl || ''), '_blank')
            }
            onSuccessCallback('null')
            break
        }

        case 'saveLastSearch': {
            savedSearch = request.arguments
            localStorage.setItem('savedSearch', JSON.stringify(savedSearch))
            onSuccessCallback('null')
            break
        }

        case 'loadLastSearch': {
            onSuccessCallback(JSON.stringify(savedSearch))
            break
        }

        case 'indicateFinishedLoading': {
            setOverlay(null)
            onSuccessCallback('null')
            break
        }

        case 'windowClose': {
            console.log('Closing window')
            onSuccessCallback('null')
            break
        }

        default: {
            // noinspection UnnecessaryLocalVariableJS
            const exhaustiveCheck: never = action
            onFailureCallback(2, `Unknown action: ${exhaustiveCheck as string}`)
        }
    }
}

export function setDarkMode(value: boolean): void {
    isDarkTheme = value
}

function setOverlay(content: string | null): void {
    webviewOverlay.style.visibility = content !== null ? 'visible' : 'hidden'
    webviewOverlay.innerHTML = content || ''
}

function escapeHTML(unsafe: string): string {
    return unsafe.replaceAll(
        // eslint-disable-next-line no-control-regex
        /[\u0000-\u002F\u003A-\u0040\u005B-\u0060\u007B-\u00FF]/g,

        char => '&#' + ('000' + char.charCodeAt(0)).slice(-4) + ';'
    )
}
