import { useLocation } from 'react-router-dom'

import { Link } from '@sourcegraph/wildcard'

import { CreateGitHubAppPage } from '../../../components/gitHubApps/CreateGitHubAppPage'

interface Props {
    defaultEvents: string[]
    defaultPermissions: Record<string, string>
}

export const BatchChangesCreateGitHubAppPage: React.FunctionComponent<Props> = ({
    defaultEvents,
    defaultPermissions,
}) => {
    const location = useLocation()
    const baseURL = new URLSearchParams(location.search).get('baseURL')
    console.log({ baseURL })
    return (
        <CreateGitHubAppPage
            defaultEvents={defaultEvents}
            defaultPermissions={defaultPermissions}
            pageTitle="Create GitHub App for commit signing"
            headerDescription={
                <>
                    Register a GitHub App to enable Sourcegraph to sign commits for Batch Change changesets on your
                    behalf.
                    {/* TODO: Update me */}
                    <Link to="/help/admin/external_service/github#using-a-github-app" className="ml-1">
                        See how GitHub App configuration works.
                    </Link>
                </>
            }
            defaultAppName="Sourcegraph Commit Signing"
            baseURL={baseURL?.length ? baseURL : undefined}
        />
    )
}
