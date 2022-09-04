import { useState } from 'react'

import { useScript } from './useScript'

export const useGoogleAPIScript = (): boolean => {
    const [ready, setReady] = useState(false)
    useScript(
        'https://apis.google.com/js/api.js',
        () => {
            if (window.gapi) {
                window.gapi.load('client', () => setReady(true))
            }
        },
        error => {
            console.error('Error loading Google APIs:', error)
        }
    )
    return ready
}

// TODO(sqs): remove
/* export const useGoogleAPI = (jwt: string, clientID: string) => {
	const [gapiClient, setGapiClient] = useState<typeof window.gapi.client>()

	useScript(
			'https://apis.google.com/js/api.js',
			() => {
					if (window.gapi) {
							void (async () => {
									await window.gapi.client.init({
											apiKey: jwt,
											clientId: clientID,
											discoveryDocs: ['https://tasks.googleapis.com/$discovery/rest'],
									})
									setGapiClient(window.gapi.client)
							})()
					}
			},
			error => {
					console.error('Error loading Google APIs:', error)
			}
	)

	return gapiClient
}
 */
