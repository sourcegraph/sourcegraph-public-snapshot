import { CompletionCallbacks } from '../client/types'
import { conversationStarter, createRequestBody, humanInput, isLoggedin, sendEvents } from '../client/utils'

// Create context menu (right click menu)
chrome.runtime.onInstalled.addListener(async () => {
    chrome.contextMenus.create({
        id: 'explain',
        title: 'Explain selected code',
        type: 'normal',
        contexts: ['selection'],
    })
    chrome.contextMenus.create({
        id: 'optimize',
        title: 'Optimize selected code',
        type: 'normal',
        contexts: ['selection'],
    })
    chrome.contextMenus.create({
        id: 'debug',
        title: 'Debug selected code',
        type: 'normal',
        contexts: ['selection'],
    })
    chrome.contextMenus.create({
        id: 'ask',
        title: 'What is wrong with the selected code?',
        type: 'normal',
        contexts: ['selection'],
    })
})

chrome.storage.local.get(['sgCodyEndpoint', 'sgCodyToken', 'sgCodyAuth'], (data: any) => {
    let endpoint = data?.sgCodyEndpoint
    let accessToken = data?.sgCodyToken
    let authStatus = data?.sgCodyAuth
    // listen to storage keys change
    chrome.storage.onChanged.addListener(function (changes, namespace) {
        for (const key in changes) {
            const storageChange = changes[key]
            switch (key) {
                case 'sgCodyEndpoint':
                    endpoint = storageChange.newValue
                    break
                case 'sgCodyToken':
                    accessToken = storageChange.newValue
                    break
                case 'sgCodyAuth':
                    authStatus = storageChange.newValue
                    break
            }
        }
    })
    // Listen to contextMenus actions
    chrome.contextMenus.onClicked.addListener(async (item, tab) => {
        // Action info returned by browser
        const code = item.selectionText
        const action = item.menuItemId
        const tId = tab?.id
        // if the tab id is presented...
        if (tId && (action === 'ask' || action === 'optimize' || action === 'explain' || action === 'debug')) {
            const isAuthed = await isLoggedin(endpoint, accessToken)
            if (!authStatus && !isAuthed) {
                // Send message to UI to ask users to sign in
                chrome.tabs.sendMessage(tId, {
                    type: 'cody',
                    data: { query: 'auth', message: false },
                })
                return
            }
            // For auth users, show "Code is typing..."
            chrome.tabs.sendMessage(tId, {
                type: 'cody',
                data: { query: 'wait', message: code },
            })
            // Create fetch request init
            const starter = action === 'ask' ? 'Tell me what is wrong with' : action
            const prompt = `${starter} the following code: ${code}`
            const groupedMsgs = [...conversationStarter, ...humanInput(prompt)]
            const fetchRequest = {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    credentials: 'include',
                    Authorization: `token ${accessToken}`,
                },
                body: JSON.stringify(createRequestBody(groupedMsgs)),
            }
            // Make fetch request
            const sgURL = new URL('/.api/completions/stream', endpoint).href
            fetch(sgURL, fetchRequest)
                .then(async (response: any) => {
                    if (!response) return
                    if (!response.ok) {
                        throw new Error(`Request failed with status ${response.status}`)
                    }
                    const stream = await response.text()
                    return stream
                })
                .then((body: any) => {
                    if (!body) return
                    const cb: CompletionCallbacks = {
                        onChange: (text: string) => {
                            chrome.tabs.sendMessage(tId, {
                                type: 'cody',
                                data: { query: 'answer', message: text.replace('\n\n', '') },
                            })
                        },
                        onComplete: () => console.log('completed'),
                        onError: (err: any) => {
                            if (err.name === 'AbortError') {
                                console.error('Request aborted:', err)
                                return
                            }
                            console.error('Request failed', err)
                        },
                    }
                    sendEvents(body, cb)
                })
                .catch((e: any) => {
                    console.log(e.message)
                })
        }
    })
})
