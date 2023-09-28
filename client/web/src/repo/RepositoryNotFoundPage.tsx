import * as React from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'

import { Code, Link, Text } from '@sourcegraph/wildcard'

import { HeroPage } from '../components/HeroPage'
import { eventLogger } from '../tracking/eventLogger'

import styles from './RepositoryNotFoundPage.module.scss'

interface Props {
    /** The name of the repository. */
    repo: string

    /** Whether the viewer is a site admin. */
    viewerCanAdminister: boolean
}

/**
 * A page informing the user that an error occurred while trying to display the repository. It
 * attempts to present the user with actions to solve the problem.
 */
export const RepositoryNotFoundPage: React.FunctionComponent<Props> = ({ repo, viewerCanAdminister }) => {
    React.useEffect(() => eventLogger.logViewEvent('RepositoryError'), [])

    return (
        <HeroPage
            icon={MapSearchIcon}
            title="Repository not found"
            subtitle={
                <div className={styles.repositoryNotFoundPage}>
                    {viewerCanAdminister && (
                        <Text>
                            As a site admin, you can add <Code>{repo}</Code> to Sourcegraph to allow users to search and
                            view it by <Link to="/site-admin/external-services">connecting an external service</Link>{' '}
                            referencing it.
                        </Text>
                    )}
                    {!viewerCanAdminister && <Text>To access this repository, contact the Sourcegraph admin.</Text>}
                </div>
            }
        />
    )
}
