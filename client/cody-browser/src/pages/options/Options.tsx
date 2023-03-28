import { useCallback, useEffect, useState } from 'react'

import '@pages/options/Options.css'

import { isLoggedin } from '../client/utils'

export default function Options(): JSX.Element {
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
    }, [accessToken, sgEndpoint])

    return (
        <div className="container mx-auto p-4">
            <div className="flex items-center flex-col mb-5 p-1">
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
                    <button className="mt-5 rounded-full bg-violet-600" form="signin" onClick={() => onSignInClick()}>
                        {saveClicked ? 'Saved!' : 'Save Settings'}
                    </button>
                </div>
            </div>
        </div>
    )
}
