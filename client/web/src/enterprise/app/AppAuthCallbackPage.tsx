import { useEffect, useRef, useState } from 'react'

import * as jsonc from 'jsonc-parser'
import { useSearchParams } from 'react-router-dom'

import { Container, ErrorAlert, Text, Link } from '@sourcegraph/wildcard'

import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { fetchSite, reloadSite, updateSiteConfiguration } from '../../site-admin/backend'
import { eventLogger } from '../../tracking/eventLogger'

export const AppAuthCallbackPage: React.FC = () => {
    useEffect(() => eventLogger.logPageView('AppAuthCallbackPage'), [])

    const [error, setError] = useState<Error | string | null>(null)

    const [searchParams] = useSearchParams()
    const code = searchParams.get('code')
    const destination = searchParams.get('destination')

    const isInvalidUrl = !code || code === ''

    const didSaveRef = useRef(false)
    useEffect(() => {
        if (isInvalidUrl) {
            return
        }
        if (didSaveRef.current) {
            return
        }
        didSaveRef.current = true

        saveAccessToken(code).catch(setError)
    }, [code, isInvalidUrl])

    return (
        <Page>
            <PageTitle title="Auth callback" />
            {isInvalidUrl || error ? (
                <ErrorAlert error={isInvalidUrl ? 'Invalid redirect URL. Please try again.' : error} />
            ) : (
                <Container>
                    <Text>Thank you for connecting your Sourcegraph.com account.</Text>

                    <Text>
                        If you are not redirected shortly, <Link to={destination ?? '/'}>continue here</Link>.
                    </Text>
                </Container>
            )}
        </Page>
    )
}

const defaultModificationOptions: jsonc.ModificationOptions = {
    formattingOptions: {
        eol: '\n',
        insertSpaces: true,
        tabSize: 2,
    },
}

async function saveAccessToken(accessToken: string): Promise<void> {
    const site = await fetchSite().toPromise()

    const content = site.configuration.effectiveContents
    const id = site.id

    const modification = jsonc.modify(content, ['app', 'accessToken'], accessToken, defaultModificationOptions)
    const modifiedContent = jsonc.applyEdits(content, modification)

    // await updateSiteConfiguration(lastConfigurationID, newContents).toPromise<boolean>()

    console.log({ content, id, modifiedContent })
}
