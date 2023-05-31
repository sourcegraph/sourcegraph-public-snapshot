import { useEffect } from 'react'

import { gql, useQuery } from '@sourcegraph/http-client'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { Button, Link, Select, Text, useLocalStorage } from '@sourcegraph/wildcard'

import { tauriInvoke } from '../app/tauriIcpUtils'
import { HeroPage } from '../components/HeroPage'
import { GetReposForCodyResult, GetReposForCodyVariables } from '../graphql-operations'

import { CodyLogo } from './components/CodyLogo'
import { CodySidebar } from './sidebar'
import { useCodySidebar, CodySidebarStoreProvider } from './sidebar/Provider'

import styles from './CodyStandalonePage.module.scss'

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
    const [selectedRepo, setSelectedRepo] = useTemporarySetting('app.codyStandalonePage.selectedRepo', '')

    const { scope, setScope, loaded, isCodyEnabled } = useCodySidebar()

    const enabled = appSetupFinished && isCodyEnabled.chat && !isCodyEnabled.needsEmailVerification
    const disabledReason: CodyDisabledReason = !appSetupFinished
        ? 'setupNotCompleted'
        : !isCodyEnabled.chat
        ? 'accountNotConnected'
        : 'emailNotVerified'

    useEffect(() => {
        if (loaded && scope.type === 'Automatic' && !scope.repositories.find(name => name === selectedRepo)) {
            setScope({ ...scope, repositories: selectedRepo ? [selectedRepo] : [] })
        }
    }, [loaded, scope, selectedRepo, setScope])

    const repoSelector = (
        <Select
            isCustomStyle={true}
            className={styles.repoSelect}
            selectSize="sm"
            label="Repo:"
            id="repo-select"
            value={selectedRepo || 'none'}
            onChange={(event: React.ChangeEvent<HTMLSelectElement>): void => {
                setSelectedRepo(event.target.value)
            }}
        >
            <option value="" disabled={true}>
                Select a repo
            </option>
            {repos.map(({ name }: { name: string }) => (
                <option key={name} value={name}>
                    {name}
                </option>
            ))}
        </Select>
    )

    return enabled ? (
        <div className="d-flex flex-column w-100">
            <CodySidebar titleContent={repoSelector} />
        </div>
    ) : (
        <CodyDisabledNotice reason={disabledReason} />
    )
}

export const CodyStandalonePage: React.FunctionComponent<{}> = () => {
    const { data } = useQuery<GetReposForCodyResult, GetReposForCodyVariables>(REPOS_QUERY, {})

    return (
        <CodySidebarStoreProvider>
            <CodyStandalonePageContext repos={data?.repositories?.nodes || []} />
        </CodySidebarStoreProvider>
    )
}
