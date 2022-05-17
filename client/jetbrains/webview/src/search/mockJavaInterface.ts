import { RequestToJava } from './jsToJavaBridgeUtil'

export function callJava(request: RequestToJava): Promise<object> {
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
    request: RequestToJava,
    onSuccessCallback: (responseAsString: string) => void,
    onFailureCallback: (errorCode: number, errorMessage: string) => void
): void {
    if (request.action === 'getTheme') {
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
    } else if (request.action === 'preview') {
        const { path } = request.arguments as { path: string }
        console.log(`Previewing "${path}"`)
    } else if (request.action === 'clearPreview') {
        console.log('Clearing preview.')
    } else if (request.action === 'open') {
        const { path } = request.arguments as { path: string }
        console.log(`Opening "${path}"`)
    } else {
        onFailureCallback(2, `Unknown action: ${request.action}`)
    }
}
