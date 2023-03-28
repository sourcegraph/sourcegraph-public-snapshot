import { useCallback, useEffect, useState } from 'react'

import logo from '@assets/img/cody.png'

import { isLoggedin } from '../client/utils'

export default function Popup(): JSX.Element {
    const [auth, setAuth] = useState(false)
    const [sgEndpoint, setSgEndpoint] = useState('https://example.sourcegraph.com')
    const [accessToken, setAccessToken] = useState('')
    const [saveClicked, setSaveClicked] = useState(false)

    // Call on mount only
    useEffect(() => {
        chrome.storage.local.get('sgCodyToken', (data: any) => {
            setAccessToken(data?.sgCodyToken)
        })
        chrome.storage.local.get('sgCodyEndpoint', (data: any) => {
            if (data?.sgCodyEndpoint) {
                setSgEndpoint(data?.sgCodyEndpoint)
            }
        })
    }, [accessToken, sgEndpoint])

    useEffect(() => {
        if (!auth && sgEndpoint && accessToken) {
            isLoggedin(sgEndpoint, accessToken).then((authState: boolean) => {
                chrome.storage.local.set({ sgCodyToken: accessToken })
                chrome.storage.local.set({ sgCodyEndpoint: sgEndpoint })
                chrome.storage.local.set({ sgCodyAuth: authState })
                setAuth(authState)
            })
        }
    }, [accessToken, sgEndpoint, auth, saveClicked])

    const onSignInClick = useCallback(async () => {
        setSaveClicked(true)
        chrome.storage.local.set({ sgCodyToken: accessToken })
        chrome.storage.local.set({ sgCodyEndpoint: sgEndpoint })
        const authUser = await isLoggedin(accessToken, sgEndpoint)
        setAuth(authUser)
    }, [accessToken, sgEndpoint])

    const onAskCodyClick = useCallback(() => {
        // Define the properties of the new popup window
        // Create the new popup window
        chrome.windows.create({
            type: 'popup',
            width: 400,
            height: 600,
            url: 'src/pages/newtab/index.html', // URL of the HTML file to be displayed in the popup
        })
    }, [])

    return (
        <div className="container mx-auto p-4 ">
            <div className="flex items-center space-x-4">
                <div className="">
                    <a href="https://docs.sourcegraph.com/cody" target="_blank" rel="noreferrer">
                        <img src={logo} className="max-h-10" alt="Cody logo" />
                    </a>
                </div>
                <p className="text-3xl p-1 text-justify">Cody</p>
            </div>
            <div className="flex items-center flex-col mb-5 p-1">
                {!auth ? (
                    <div className="flex flex-col w-full mt-5" id="signin">
                        <p>Sourcegraph Instance URL</p>
                        <input
                            type="url"
                            value={sgEndpoint}
                            onChange={e => setSgEndpoint(e.target.value)}
                            className="w-full p-1 my-4 rounded-lg bg-gray-700 text-slate-200"
                            required={true}
                        />
                        <p>Access Token</p>
                        <input
                            type="password"
                            value={accessToken}
                            onChange={e => setAccessToken(e.target.value)}
                            className="w-full p-1 my-4 rounded-lg bg-gray-700 text-slate-200"
                            required={true}
                        />
                        {auth && <p>Status: You are currently logged in.</p>}
                        {!auth && accessToken && sgEndpoint && <p className="text-pink-500">⛔️ Failed to sign in.</p>}
                        <button
                            className="mt-5 rounded-full bg-violet-600"
                            form="signin"
                            onClick={() => onSignInClick()}
                        >
                            {saveClicked ? 'Saved!' : 'Save Settings'}
                        </button>
                    </div>
                ) : (
                    <div className="flex-col w-full mt-5 flex justify-between" id="signin">
                        <p className="text-green-500">You are signed in!</p>
                        <input
                            type="url"
                            value={sgEndpoint}
                            className="w-full p-1 my-4 rounded-lg border-2 border-green-500 bg-gray-700 text-slate-200"
                            disabled={true}
                        />
                        <p className="text-center">Start selecting code snippet</p>
                        <p className="text-center my-2">-- or --</p>
                        <p className="text-center">Start a conversation with Cody</p>
                        <button className="mt-5 rounded-full bg-violet-600" onClick={() => onAskCodyClick()}>
                            Ask Cody
                        </button>
                    </div>
                )}
            </div>
        </div>
    )
}
