import React, { useCallback, useEffect, useState } from 'react'

import { gql, useMutation } from '@apollo/client'
import { debounce } from 'lodash'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { Button, Checkbox, Container, Link, PageHeader } from '@sourcegraph/wildcard'

import { refreshAuthenticatedUser } from '../../../auth'
import { PageTitle } from '../../../components/PageTitle'
import { SetUserSearchableResult, SetUserSearchableVariables } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { UserSettingsAreaRouteContext } from '../UserSettingsArea'

import styles from './UserSettingsPrivacyPage.module.scss'

export const USER_SEARCHABLE_MUTATION = gql`
    mutation SetUserSearchable($isSearchable: Boolean!) {
        setSearchable(searchable: $isSearchable) {
            alwaysNil
        }
    }
`

interface Props extends Pick<UserSettingsAreaRouteContext, 'authenticatedUser'> {
    authenticatedUser: AuthenticatedUser
}

export const UserSettingsPrivacyPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    authenticatedUser,
}) => {
    useEffect(() => eventLogger.logViewEvent('UserProfile'), [])

    const [setUserSearchable, { loading, error }] = useMutation<SetUserSearchableResult, SetUserSearchableVariables>(
        USER_SEARCHABLE_MUTATION
    )

    const [disableSearchable, setDisableSearchable] = useState(!authenticatedUser.searchable)

    const onCheckboxChange = useCallback(() => {
        setDisableSearchable(!disableSearchable)
    }, [disableSearchable])

    const onSaveSearchableState = useCallback(async () => {
        eventLogger.log('SaveUserSearchable', !disableSearchable)
        try {
            await setUserSearchable({
                variables: {
                    isSearchable: !disableSearchable,
                },
            })
            // The edited user is the current user, immediately reflect the changes in the UI.
            // TODO: Migrate this to use the Apollo cache
            await refreshAuthenticatedUser().toPromise()
            eventLogger.log('SaveUserSearchableOK')
        } catch {
            eventLogger.log('SaveUserSearchableFailed')
        }
    }, [disableSearchable, setUserSearchable])

    const debounceSaveSearchableState = debounce(onSaveSearchableState, 500, { leading: true })
    const hasChanges = authenticatedUser.searchable === disableSearchable

    return (
        <div>
            <PageTitle title="Profile" />
            <PageHeader path={[{ text: 'Privacy' }]} headingElement="h2" className={styles.heading} />
            <Container>
                <Checkbox
                    name="userSearchable"
                    id="userSearchable"
                    value="searchable"
                    checked={disableSearchable}
                    onChange={onCheckboxChange}
                    label="Don’t share my profile in autocomplete search results on Sourcegraph Cloud"
                    message="Other Sourcegraph users will only be able to add you to organizations if they know your username or email"
                />
                <div className="d-flex justify-content-start mt-4 mb-3 border-bottom">
                    <Button
                        type="button"
                        className="mb-3"
                        variant="primary"
                        onClick={debounceSaveSearchableState}
                        disabled={loading || !hasChanges}
                    >
                        Save
                    </Button>
                </div>
                <div className="d-flex justify-content-start mb-3">
                    <p>
                        Learn more about{' '}
                        <Link to="https://docs.sourcegraph.com/code_search/explanations/sourcegraph_cloud#how-secure-is-sourcegraph-cloud-can-sourcegraph-see-my-code">
                            how your data and privacy is protected on Sourcegraph Cloud.
                        </Link>
                    </p>
                </div>
                {error && <ErrorAlert className="mt-2" error={error} />}
            </Container>
        </div>
    )
}
