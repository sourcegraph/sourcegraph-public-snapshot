import React from 'react'

import { mdiAccount, mdiUpload } from '@mdi/js'
import { Navigate } from 'react-router-dom'

import { LoadingSpinner, PageHeader, Icon, H1, Text, H3, Link, Button } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import { RepositoryFields } from '../../graphql-operations'

import styles from './RepositoryOwnPage.module.scss'

/**
 * Properties passed to all page components in the repository code navigation area.
 */
export interface RepositoryOwnAreaPageProps extends BreadcrumbSetters {
    /** The active repository. */
    repo: RepositoryFields
    authenticatedUser: AuthenticatedUser | null
}
const BREADCRUMB = { key: 'own', element: 'Ownership' }

// Render the name of the CODEOWNERS file in CamelCase in a span with the text-uppercase class
// so that screen readers will read it correctly.
const CodeOwnersName = React.memo(function CodeOwnersName() {
    return <span className="text-uppercase">CodeOwners</span>
})

export const RepositoryOwnPage: React.FunctionComponent<RepositoryOwnAreaPageProps> = ({
    useBreadcrumb,
    repo,
    authenticatedUser,
}) => {
    useBreadcrumb(BREADCRUMB)

    const [ownEnabled, status] = useFeatureFlag('search-ownership')

    if (status === 'initial') {
        return (
            <div className="container d-flex justify-content-center mt-3">
                <LoadingSpinner />
            </div>
        )
    }

    if (!ownEnabled) {
        return <Navigate to={repo.url} replace={true} />
    }

    return (
        <Page>
            <PageTitle title="Sourcegraph Own" />
            <PageHeader
                description={
                    <>
                        Sourcegraph Own can provide code ownership data for this repository via an upload or a committed{' '}
                        <CodeOwnersName /> file. <Link to="/help/own">Learn more</Link>
                    </>
                }
            >
                <H1 as="h2">
                    <Icon svgPath={mdiAccount} aria-hidden={true} />
                    <span className="ml-2">Ownership</span>
                </H1>
            </PageHeader>

            <div className={styles.columns}>
                <div>
                    <H3>
                        Upload a <CodeOwnersName /> file
                    </H3>
                    <Text>
                        Each owner must be either a Sourcegraph username, a Sourcegraph team name, or an email address.
                    </Text>

                    <Button variant="primary">
                        <Icon svgPath={mdiUpload} aria-hidden={true} className="mr-2" />
                        Upload file
                    </Button>
                </div>

                <div className={styles.or}>
                    <div className={styles.orLine} />
                    <div className="py-2">or</div>
                    <div className={styles.orLine} />
                </div>

                <div>
                    <H3>
                        Commit a <CodeOwnersName /> file
                    </H3>
                    <Text>
                        A <CodeOwnersName /> file living in the root of your repository. Owners must be{' '}
                        {getCodeHostName(repo)} usernames or email addresses.
                    </Text>
                </div>
            </div>
        </Page>
    )
}

const getCodeHostName = (repo: RepositoryFields): string => {
    const externalServiceKind = repo.externalURLs[0]?.serviceKind

    switch (externalServiceKind) {
        case 'GITHUB':
            return 'GitHub'
        case 'GITLAB':
            return 'GitLab'
        default:
            return 'code host'
    }
}
