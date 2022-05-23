import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { Search } from '../search/App'
import { Request } from '../search/jsToJavaBridgeUtil'

let savedSearch: Search = {
    query: '',
    caseSensitive: false,
    patternType: SearchPatternType.literal,
    selectedSearchContextSpec: 'global',
}

const codeDetailsNode = document.querySelector('#code-details') as HTMLPreElement
const iframeNode = document.querySelector('#webview') as HTMLIFrameElement

function callJava(request: Request): Promise<object> {
    return new Promise((resolve, reject) => {
        const requestAsString = JSON.stringify(request)
        const onSuccessCallback = (responseAsString: string): void => {
            resolve(JSON.parse(responseAsString))
        }
        const onFailureCallback = (errorCode: number, errorMessage: string): void => {
            reject(new Error(`${errorCode} - ${errorMessage}`))
        }
        console.log(`Got this request: ${requestAsString}`)
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
                    instanceURL: 'https://sourcegraph.com',
                    isGlobbingEnabled: true,
                    accessToken: null,
                })
            )
            break
        }

        case 'getTheme': {
            onSuccessCallback(
                JSON.stringify({
                    isDarkTheme: true,
                    backgroundColor: 'blue',
                    buttonArc: '2px',
                    buttonColor: 'red',
                    color: 'green',
                    font: 'Times New Roman',
                    fontSize: '12px',
                    labelBackground: 'gray',
                })
            )
            break
        }

        case 'preview': {
            const { content, absoluteOffsetAndLengths } = request.arguments

            const start = absoluteOffsetAndLengths[0][0]
            const length = absoluteOffsetAndLengths[0][1]

            let htmlContent: string = escapeHTML(content.slice(0, start))
            htmlContent += `<span id="code-details-highlight">${escapeHTML(
                content.slice(start, start + length)
            )}</span>`
            htmlContent += escapeHTML(content.slice(start + length))

            codeDetailsNode.innerHTML = htmlContent

            document.querySelector('#code-details-highlight')?.scrollIntoView({ block: 'center', inline: 'center' })

            onSuccessCallback('{}')
            break
        }

        case 'clearPreview': {
            codeDetailsNode.textContent = ''
            onSuccessCallback('{}')
            break
        }

        case 'open': {
            const { path } = request.arguments
            alert(`Opening ${path}`)
            onSuccessCallback('{}')
            break
        }

        case 'saveLastSearch': {
            savedSearch = request.arguments
            onSuccessCallback('{}')
            break
        }

        case 'loadLastSearch': {
            onSuccessCallback(JSON.stringify(savedSearch))
            break
        }

        case 'indicateFinishedLoading': {
            onSuccessCallback('{}')
            break
        }

        default: {
            const exhaustiveCheck: never = action
            onFailureCallback(2, `Unknown action: ${exhaustiveCheck as string}`)
        }
    }
}

/* Initialize app for standalone server */
iframeNode.addEventListener('load', () => {
    const iframeWindow = iframeNode.contentWindow
    if (iframeWindow !== null) {
        iframeWindow.callJava = callJava
        iframeWindow.initializeSourcegraph()
    }
})

function escapeHTML(unsafe: string): string {
    return unsafe.replace(
        /[\u0000-\u002F\u003A-\u0040\u005B-\u0060\u007B-\u00FF]/g,
        c => '&#' + ('000' + c.charCodeAt(0)).slice(-4) + ';'
    )
}
