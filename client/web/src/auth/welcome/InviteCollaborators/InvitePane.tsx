import classNames from 'classnames'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import React, { useCallback, useMemo, useState } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import { CopyableText } from '@sourcegraph/web/src/components/CopyableText'
import { LoadingSpinner, Button } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'
import { LoaderButton } from '../../../components/LoaderButton'
import { eventLogger } from '../../../tracking/eventLogger'
import { UserAvatar } from '../../../user/UserAvatar'

import { InvitableCollaborator } from './InviteCollaborators'
import styles from './InviteCollaborators.module.scss'
import { useInviteEmailToSourcegraph } from './useInviteEmailToSourcegraph'

interface Props {
    user: AuthenticatedUser
    className?: string
    invitableCollaborators: InvitableCollaborator[]
    isLoadingCollaborators: boolean
}

export const InvitePane: React.FunctionComponent<Props> = ({
    user,
    className,
    invitableCollaborators,
    isLoadingCollaborators,
}) => {
    const inviteEmailToSourcegraph = useInviteEmailToSourcegraph()
    const preventSubmit = useCallback((event: React.FormEvent<HTMLFormElement>): void => event.preventDefault(), [])
    const [query, setQuery] = useState('')

    const filteredCollaborators: InvitableCollaborator[] = useMemo(() => {
        if (query.trim() === '') {
            return invitableCollaborators
        }
        return invitableCollaborators.filter(
            person =>
                person.name.toLocaleLowerCase().includes(query.toLocaleLowerCase()) ||
                person.email.toLocaleLowerCase().includes(query.toLocaleLowerCase())
        )
    }, [query, invitableCollaborators])

    const [inviteError, setInviteError] = useState<ErrorLike | null>(null)
    const [loadingInvites, setLoadingInvites] = useState<Set<string>>(new Set<string>())
    const [successfulInvites, setSuccessfulInvites] = useState<Set<string>>(new Set<string>())

    const inviteableCollaborators: InvitableCollaborator[] = useMemo(
        () =>
            filteredCollaborators.filter(
                person => !successfulInvites.has(person.email) && !loadingInvites.has(person.email)
            ),
        [filteredCollaborators, loadingInvites, successfulInvites]
    )

    const invitePerson = useCallback(
        async (person: InvitableCollaborator): Promise<void> => {
            if (loadingInvites.has(person.email) || successfulInvites.has(person.email)) {
                return
            }
            setLoadingInvites(set => new Set(set).add(person.email))

            try {
                await inviteEmailToSourcegraph({ variables: { email: person.email } })

                setLoadingInvites(set => {
                    const removed = new Set(set)
                    removed.delete(person.email)
                    return removed
                })
                setSuccessfulInvites(set => new Set(set).add(person.email))

                eventLogger.log('UserInvitationsSentEmailInvite')
            } catch (error) {
                setInviteError(error)
            }
        },
        [loadingInvites, successfulInvites, inviteEmailToSourcegraph]
    )
    const invitePersonClicked = useCallback(
        (person: InvitableCollaborator) => async (): Promise<void> => {
            await invitePerson(person)
        },
        [invitePerson]
    )
    const [isInvitingAll, setIsInvitingAll] = useState(false)
    const inviteAllClicked = useCallback(async (): Promise<void> => {
        setIsInvitingAll(true)
        for (const person of inviteableCollaborators) {
            await invitePerson(person)
        }
        setIsInvitingAll(false)
    }, [invitePerson, inviteableCollaborators])

    const inviteURL = `${window.context.externalURL}/sign-up?invitedBy=${user.username}`
    return (
        <div className={classNames(className, 'mx-2')}>
            <div className={styles.titleDescription}>
                <h3>Introduce friends and colleagues to Sourcegraph</h3>
                <p className="text-muted mb-4">
                    We’ve selected a few collaborators you might want to invite to Sourcegraph. These users won’t be
                    able to see your code unless they have access to it on the code host and also add that code to
                    Sourcegraph.
                </p>
            </div>
            {isErrorLike(inviteError) && <ErrorAlert error={inviteError} />}
            <div className="border overflow-hidden rounded">
                <header>
                    <div className="py-3 px-3 d-flex justify-content-between align-items-center">
                        <h4 className="flex-1 m-0">Collaborators</h4>
                        <Form
                            onSubmit={preventSubmit}
                            className="flex-1 d-inline-flex justify-content-between flex-row"
                        >
                            <input
                                className="form-control"
                                type="search"
                                placeholder="Filter by email or username"
                                name="query"
                                autoComplete="off"
                                autoCorrect="off"
                                autoCapitalize="off"
                                spellCheck={false}
                                onChange={event => {
                                    setQuery(event.target.value)
                                }}
                            />
                        </Form>
                    </div>
                </header>
                <div className={classNames('mb-3', styles.invitableCollaborators)}>
                    {!isLoadingCollaborators &&
                        filteredCollaborators.map((person, index) => (
                            <div
                                className={classNames('d-flex', 'ml-3', 'align-items-center', index !== 0 && 'mt-3')}
                                key={person.email}
                            >
                                <UserAvatar
                                    className={classNames('icon-inline', 'mr-3', styles.avatar)}
                                    user={person}
                                />
                                <div>
                                    <strong>{person.displayName}</strong>
                                    <div className="text-muted">{person.email}</div>
                                </div>
                                {loadingInvites.has(person.email) ? (
                                    <LoadingSpinner inline={true} className={classNames('ml-auto', 'mr-3')} />
                                ) : successfulInvites.has(person.email) ? (
                                    <span className="text-muted ml-auto mr-3">
                                        <CheckCircleIcon className="icon-inline mr-1" />
                                        Invited
                                    </span>
                                ) : (
                                    <Button
                                        variant="secondary"
                                        outline={true}
                                        size="sm"
                                        className={classNames('ml-auto', 'mr-3', styles.inviteButton)}
                                        onClick={invitePersonClicked(person)}
                                    >
                                        Invite
                                    </Button>
                                )}
                            </div>
                        ))}
                    {isLoadingCollaborators && (
                        <div className="text-muted d-flex justify-content-center mt-3">
                            <LoadingSpinner inline={false} />
                        </div>
                    )}
                    {!isLoadingCollaborators && filteredCollaborators.length === 0 && (
                        <div className="text-muted d-flex justify-content-center mt-3">
                            No collaborators found. Try sending them a direct link below
                        </div>
                    )}
                </div>

                <LoaderButton
                    loading={isInvitingAll}
                    variant="success"
                    className="d-block ml-auto mb-3 mr-3"
                    onClick={inviteAllClicked}
                    alwaysShowLabel={true}
                    disabled={isLoadingCollaborators || inviteableCollaborators.length === 0 || isInvitingAll}
                    label={
                        inviteableCollaborators.length === 0 && filteredCollaborators.length !== 0
                            ? `Invited ${filteredCollaborators.length} users`
                            : `${isInvitingAll ? 'Inviting' : 'Invite'} ${
                                  isLoadingCollaborators || inviteableCollaborators.length === 0
                                      ? ''
                                      : `${inviteableCollaborators.length} `
                              }users`
                    }
                />
            </div>
            <div>
                <header>
                    <div className="py-3 d-flex justify-content-between align-items-center">
                        <h4 className="m-0">Or invite by sending a link</h4>
                    </div>
                </header>
                <CopyableText
                    className="mb-3 flex-1"
                    text={inviteURL}
                    flex={true}
                    size={inviteURL.length}
                    onCopy={() => eventLogger.log('UserInvitationsCopiedInviteLink')}
                />
            </div>
        </div>
    )
}
