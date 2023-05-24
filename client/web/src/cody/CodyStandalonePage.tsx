import { gql, useQuery } from '@sourcegraph/http-client'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { Button, Link, Select, Text, useLocalStorage } from '@sourcegraph/wildcard'

import { tauriInvoke } from '../app/tauriIcpUtils'
import { HeroPage } from '../components/HeroPage'
import { GetReposForCodyResult, GetReposForCodyVariables } from '../graphql-operations'

import { CodyLogo } from './components/CodyLogo'
import { CodySidebar } from './sidebar/CodySidebar'
import { useChatStore } from './stores/chat'
import { useIsCodyEnabled } from './useIsCodyEnabled'

import styles from './CodyStandalonePage.module.scss'

const noop = (): void => {}

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

export const CodyStandalonePage: React.FunctionComponent<{}> = () => {
    // eslint-disable-next-line no-restricted-syntax
    const [appSetupFinished] = useLocalStorage('app.setup.finished', false)
    const { chat, needsEmailVerification } = useIsCodyEnabled()

    const isCodyEnabled = appSetupFinished && chat && !needsEmailVerification
    const disabledReason: CodyDisabledReason = !appSetupFinished
        ? 'setupNotCompleted'
        : !chat
        ? 'accountNotConnected'
        : 'emailNotVerified'

    return isCodyEnabled ? <CodyChat /> : <CodyDisabledNotice reason={disabledReason} />
}

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

const CodyChat: React.FunctionComponent<{}> = () => {
    const { data } = useQuery<GetReposForCodyResult, GetReposForCodyVariables>(REPOS_QUERY, {})

    const [selectedRepo, setSelectedRepo] = useTemporarySetting('app.codyStandalonePage.selectedRepo', '')
    useChatStore({ codebase: selectedRepo || '', setIsCodySidebarOpen: noop })

    const repos = data?.repositories.nodes ?? []

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
            {repos.map(({ name }) => (
                <option key={name} value={name}>
                    {name}
                </option>
            ))}
        </Select>
    )

    return (
        <div className="d-flex flex-column w-100">
            <CodySidebar titleContent={repoSelector} />
        </div>
    )
}
