import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { Search } from './App'
import { Request } from './jsToJavaBridgeUtil'

let savedSearch: Search = {
    query: '',
    caseSensitive: false,
    patternType: SearchPatternType.literal,
    selectedSearchContextSpec: 'global',
}

export function callJava(request: Request): Promise<object> {
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
            const { path } = request.arguments
            console.log(`Previewing "${path}"`)
            onSuccessCallback('{}')
            break
        }

        case 'clearPreview': {
            console.log('Clearing preview.')
            onSuccessCallback('{}')
            break
        }

        case 'open': {
            const { path } = request.arguments
            console.log(`Opening "${path}"`)
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
