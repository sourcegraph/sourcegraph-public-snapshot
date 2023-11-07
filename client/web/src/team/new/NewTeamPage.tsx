import React, { useCallback, useEffect, useState } from 'react'

import { useLocation, useNavigate } from 'react-router-dom'

import { Container, PageHeader, Link, Input, ErrorAlert, Form, Label } from '@sourcegraph/wildcard'

import { TEAM_DISPLAY_NAME_MAX_LENGTH, TEAM_NAME_MAX_LENGTH, VALID_TEAM_NAME_REGEXP } from '..'
import { LoaderButton } from '../../components/LoaderButton'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { useCreateTeam } from '../backend'

import { ParentTeamSelect } from './team-select/ParentTeamSelect'

import styles from './NewTeamPage.module.scss'

export interface NewTeamPageProps {}

export const NewTeamPage: React.FunctionComponent<React.PropsWithChildren<NewTeamPageProps>> = () => {
    const navigate = useNavigate()
    const location = useLocation()

    const [name, setName] = useState<string>('')
    const [displayName, setDisplayName] = useState<string>('')
    const [parentTeam, setParentTeam] = useState<string | null>(() => {
        const query = new URLSearchParams(location.search)
        if (query.has('parentTeam')) {
            return query.get('parentTeam')
        }
        return null
    })

    const onNameChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        const hyphenatedName = event.currentTarget.value.replaceAll(/\s/g, '-')
        setName(hyphenatedName)
    }

    const onDisplayNameChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        setDisplayName(event.currentTarget.value)
    }

    const [createTeam, { loading: createLoading, data: createTeamResult, error: createTeamError }] = useCreateTeam()

    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        event => {
            event.preventDefault()

            if (!event.currentTarget.checkValidity()) {
                return
            }

            createTeam({
                variables: {
                    name,
                    displayName: displayName !== '' ? displayName : null,
                    parentTeam: parentTeam !== '' ? parentTeam : null,
                },
            }).catch(error => {
                // eslint-disable-next-line no-console
                console.error(error)
            })
        },
        [displayName, name, createTeam, parentTeam]
    )

    useEffect(() => {
        if (createTeamResult) {
            navigate(createTeamResult.createTeam.url)
        }
    }, [createTeamResult, navigate])

    return (
        <Page className={styles.newTeamPage}>
            <PageTitle title="New team" />
            <PageHeader
                path={[{ text: 'Create a new team' }]}
                description={
                    <>
                        A team is a group of users. See <Link to="/help/admin/teams">Teams documentation</Link> for
                        information about configuring teams.
                    </>
                }
                className="mb-3"
            />
            <Form onSubmit={onSubmit}>
                <Container className="mb-3">
                    {createTeamError && <ErrorAlert className="mb-3" error={createTeamError} />}

                    <Label htmlFor="new-team--name">Team name</Label>
                    <Input
                        id="new-team--name"
                        placeholder="engineering"
                        pattern={VALID_TEAM_NAME_REGEXP}
                        maxLength={TEAM_NAME_MAX_LENGTH}
                        required={true}
                        autoCorrect="off"
                        autoComplete="off"
                        autoFocus={true}
                        value={name}
                        onChange={onNameChange}
                        disabled={createLoading}
                        message="A team name consists of letters, numbers, hyphens (-), dots (.) and may not begin
                        or end with a dot, nor begin with a hyphen."
                    />

                    <Label htmlFor="new-team--displayname" className="mt-2">
                        Display name
                    </Label>
                    <Input
                        id="new-team--displayname"
                        placeholder="Engineering Team"
                        maxLength={TEAM_DISPLAY_NAME_MAX_LENGTH}
                        autoCorrect="off"
                        value={displayName}
                        onChange={onDisplayNameChange}
                        disabled={createLoading}
                    />

                    <Label htmlFor="new-team--parentteam" className="mt-2">
                        Parent team
                    </Label>
                    <ParentTeamSelect
                        id="new-team--parentteam"
                        initial={parentTeam ?? undefined}
                        disabled={createLoading}
                        onSelect={setParentTeam}
                    />
                </Container>

                <LoaderButton
                    type="submit"
                    loading={createLoading}
                    disabled={createLoading}
                    variant="primary"
                    alwaysShowLabel={true}
                    label="Create team"
                />
            </Form>
        </Page>
    )
}
