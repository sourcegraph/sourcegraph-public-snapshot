import '@pages/newtab/Newtab.css'

import { useCallback, useEffect, useRef, useState } from 'react'

import logo from '@assets/img/cody.png'
import ReactMarkdown from 'react-markdown'

import { chromeStorageData, CompletionCallbacks, Message } from '../client/types'
import { conversationStarter, createRequestBody, humanInput, isLoggedin, sendEvents } from '../client/utils'

export default function App(): JSX.Element {
    const [auth, setAuth] = useState(false)
    const [codyIsTyping, setCodyIsTyping] = useState(false)
    const [sgEndpoint, setSgEndpoint] = useState('https://example.sourcegraph.com')
    const [accessToken, setAccessToken] = useState('')
    const [storedValue, setStoredValue] = useState('')
    const [input, setInput] = useState('')
    const [messages, setMessages] = useState<Message[]>([])
    const chatboxRef = useRef<HTMLInputElement>(null)

    useEffect(() => {
        chatboxRef.current?.scrollIntoView?.({ behavior: 'smooth' })
    }, [messages, chatboxRef.current])

    // Call on mount only
    useEffect(() => {
        chrome.storage.local.get(['sgCodyEndpoint', 'sgCodyToken', 'sgCodyAuth'], (data: chromeStorageData) => {
            if (data?.sgCodyEndpoint) {
                setSgEndpoint(data?.sgCodyEndpoint)
            }
            if (data?.sgCodyToken) {
                setAccessToken(data?.sgCodyToken)
            }
            if (data?.sgCodyEndpoint && data?.sgCodyToken) {
                isLoggedin(data?.sgCodyEndpoint, data?.sgCodyToken).then(status => {
                    chrome.storage.local.set({ sgCodyAuth: status })
                    setAuth(status)
                })
            }
        })
    }, [])

    useEffect(() => {
        chrome.storage.local.get('sgCodyAuth', data => {
            setAuth(data?.sgCodyAuth)
        })
    }, [accessToken, sgEndpoint, auth])

    const onChatKeyDown = (event: React.KeyboardEvent<HTMLTextAreaElement>): void => {
        if (event.key === 'Enter' && !event.shiftKey && input) {
            event.preventDefault()
            event.stopPropagation()
            onChatSubmit()
        }
    }

    const onSignInClick = useCallback(async () => {
        if (!accessToken || !sgEndpoint) {
            return
        }
        chrome.storage.local.set({ sgCodyToken: accessToken })
        chrome.storage.local.set({ sgCodyEndpoint: sgEndpoint })
        const authStatus = await isLoggedin(sgEndpoint, accessToken)
        setAuth(authStatus)
    }, [accessToken, sgEndpoint])

    const onChatSubmit = useCallback(() => {
        console.log('sending request to cody')
        if (!input) return
        const groupedMsgs = [...messages, ...humanInput(input)]
        setMessages(groupedMsgs)

        // create request
        const request = {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                Authorization: `token ${accessToken}`,
            },
            body: JSON.stringify(createRequestBody([...conversationStarter, ...groupedMsgs])),
        }
        setCodyIsTyping(true)
        setInput('')
        // make request
        const sgURL = new URL('/.api/completions/stream', sgEndpoint).href
        fetch(sgURL, request)
            .then(async response => {
                if (!response) return
                if (!response.ok) {
                    throw new Error(`Request failed with status ${response.status}`)
                }
                const stream = await response.text()
                return stream
            })
            .then(body => {
                if (!body) return
                const cb: CompletionCallbacks = {
                    onChange: (text: string) => {
                        const botResponse = { speaker: 'assistant', text: text } as Message
                        groupedMsgs[groupedMsgs.length - 1] = botResponse
                        setMessages(groupedMsgs)
                        setStoredValue(text)
                        setCodyIsTyping(true)
                    },
                    onComplete: () => {
                        setCodyIsTyping(false)
                        console.log('completed')
                    },
                    onError: (err: any) => {
                        setCodyIsTyping(false)
                        console.error(err)
                    },
                }
                sendEvents(body, cb)
            })
            .catch(e => {
                console.error(e.message)
                setCodyIsTyping(false)
            })
    }, [input, accessToken, storedValue])

    return (
        <div className="cody-root container mx-auto p-4">
            <div className="flex items-center space-x-4">
                <div className="">
                    <img src={logo} className="max-h-10" alt="Cody logo" />
                </div>
                <p className="text-3xl p-1 text-justify">Cody</p>
            </div>

            <div className="flex flex-col mb-5 p-1 h-full">
                {!auth && (
                    <div className="flex flex-col w-full mt-5" id="signin">
                        <p>Sourcegraph Instance URL</p>
                        <input
                            type="url"
                            value={sgEndpoint}
                            onChange={e => setSgEndpoint(e.target.value)}
                            className="w-full p-1 my-4 rounded-lg"
                            required={true}
                        />
                        <p>Access Token</p>
                        <input
                            type="password"
                            value={accessToken}
                            onChange={e => setAccessToken(e.target.value)}
                            className="w-full p-1 my-4 rounded-lg"
                            required={true}
                        />
                        <button
                            className="mt-10 rounded-full bg-violet-600"
                            form="signin"
                            onClick={() => onSignInClick()}
                        >
                            Sign in
                        </button>
                    </div>
                )}
                {auth && (
                    <div className="flex justify-between flex-col mt-2 p-1 h-full overflow-hidden">
                        <div className="flex justify-start flex-col p-2 h-full overflow-auto">
                            {!messages.length && (
                                <p className="rounded p-5 bg-blue-500">Ask Cody any programming questions!</p>
                            )}
                            {messages?.map((msg, index) => (
                                <div key={index} className="flex-col mb-5">
                                    <p className="flex-col mb-1 font-bold">
                                        {msg.speaker === 'human' ? 'You' : 'Cody'}
                                    </p>
                                    <div key={index} className="p-3 rounded bubble">
                                        <ReactMarkdown>{msg.text}</ReactMarkdown>
                                        {!msg.text && codyIsTyping && <p>Cody is typing...</p>}
                                    </div>
                                    <div ref={chatboxRef} />
                                </div>
                            ))}
                        </div>
                        <form id="chat" className="mt-2 p-2 flex-col justify-end" onSubmit={e => e.preventDefault()}>
                            <textarea
                                className="w-full mt-4 rounded-lg p-2 bg-gray-700 text-slate-200"
                                rows={5}
                                value={input}
                                onChange={e => setInput(e.target.value)}
                                onKeyDown={onChatKeyDown}
                            />
                            <button
                                className="mt-0 p-2 rounded-lg w-full bg-violet-600"
                                type="submit"
                                form="chat"
                                onClick={() => onChatSubmit()}
                            >
                                Send
                            </button>
                        </form>
                    </div>
                )}
            </div>
        </div>
    )
}
