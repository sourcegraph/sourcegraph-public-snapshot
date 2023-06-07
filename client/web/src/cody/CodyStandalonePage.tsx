import { useState, useEffect } from 'react'

import { gql, useQuery } from '@sourcegraph/http-client'
import { Button, Link, Text, useLocalStorage } from '@sourcegraph/wildcard'

import { tauriInvoke } from '../app/tauriIcpUtils'
import { HeroPage } from '../components/HeroPage'
import { GetReposForCodyResult, GetReposForCodyVariables } from '../graphql-operations'

import { CodyLogo } from './components/CodyLogo'
import { CodySidebar } from './sidebar'
import { useCodySidebar, CodySidebarStoreProvider } from './sidebar/Provider'

const REPOS_QUERY = gql`
    query GetReposForCody {
        repositories(first: 1000) {
            nodes {
                name
                embeddingExists
            }
        }
    }
`

type CodyDisabledReason = 'setupNotCompleted' | 'accountNotConnected' | 'emailNotVerified'

const reasonBodies: Record<CodyDisabledReason, () => React.ReactNode> = {
    setupNotCompleted: () => (
        <>
            <Text className="mt-3">You need to finish setting up the Sourcegraph app to use Cody.</Text>
            <Button variant="primary" size="lg" onClick={() => tauriInvoke('show_main_window')}>
                Open Setup
            </Button>
        </>
    ),
    accountNotConnected: () => (
        <>
            <Text className="mt-3">You need to connect your Sourcegraph.com account to use Cody.</Text>
            <Button
                variant="primary"
                size="lg"
                as={Link}
                to="https://sourcegraph.com/user/settings/tokens/new/callback?requestFrom=APP&destination="
                target="_blank"
            >
                Connect to Sourcegraph.com
            </Button>
        </>
    ),
    emailNotVerified: () => (
        <>
            <Text className="mt-3">
                Your Sourcegraph.com account does not have a verified email address. Please verify your email and
                restart the Sourcegraph app.
            </Text>
            <Button
                variant="primary"
                size="lg"
                as={Link}
                to="https://sourcegraph.com/user/settings/emails"
                target="_blank"
            >
                Verify Email
            </Button>
        </>
    ),
}

const CodyDisabledNotice: React.FunctionComponent<{ reason: CodyDisabledReason }> = ({ reason }) => (
    <HeroPage
        className="mx-3"
        icon={CodyLogo}
        iconClassName="pr-1" // Optically center the icon
        title="Cody is disabled"
        body={reasonBodies[reason]()}
    />
)

const CodyStandalonePageContext: React.FC<{ repos: GetReposForCodyResult['repositories']['nodes'] }> = ({ repos }) => {
    // eslint-disable-next-line no-restricted-syntax
    const [appSetupFinished] = useLocalStorage('app.setup.finished', false)
    const { scope, setScope, isCodyEnabled } = useCodySidebar()
    const [scopeInitialized, setScopeInitialized] = useState(false)

    const enabled = appSetupFinished && isCodyEnabled.chat && !isCodyEnabled.needsEmailVerification
    const disabledReason: CodyDisabledReason = !appSetupFinished
        ? 'setupNotCompleted'
        : !isCodyEnabled.chat
        ? 'accountNotConnected'
        : 'emailNotVerified'

    useEffect(() => {
        if (!scope.repositories.length && !scopeInitialized && repos.length) {
            setScope({ ...scope, repositories: [repos[0].name] })
        }

        setScopeInitialized(true)
    }, [setScope, repos, scopeInitialized, setScopeInitialized, scope])

    return enabled ? (
        <div className="d-flex flex-column w-100">
            <CodySidebar />
        </div>
    ) : (
        <CodyDisabledNotice reason={disabledReason} />
    )
}

export const CodyStandalonePage: React.FunctionComponent<{}> = () => {
    const { data } = useQuery<GetReposForCodyResult, GetReposForCodyVariables>(REPOS_QUERY, {})

    return (
        <CodySidebarStoreProvider>
            {data ? <CodyStandalonePageContext repos={data.repositories.nodes} /> : null}
        </CodySidebarStoreProvider>
    )
}
