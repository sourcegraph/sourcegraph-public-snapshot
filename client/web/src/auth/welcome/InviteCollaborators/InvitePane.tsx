import React, { useCallback, useMemo, useState } from 'react'

import classNames from 'classnames'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import { LoadingSpinner, Button, Link, Icon, Typography } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'
import { CopyableText } from '../../../components/CopyableText'
import { LoaderButton } from '../../../components/LoaderButton'
import { eventLogger } from '../../../tracking/eventLogger'
import { UserAvatar } from '../../../user/UserAvatar'
import { useSteps } from '../../Steps'

import { InvitableCollaborator } from './InviteCollaborators'
import { useInviteEmailToSourcegraph } from './useInviteEmailToSourcegraph'

import styles from './InviteCollaborators.module.scss'

const SELECT_REPOS_STEP = 2

interface Props {
    user: AuthenticatedUser
    className?: string
    invitableCollaborators: InvitableCollaborator[]
    isLoadingCollaborators: boolean
}

export const InvitePane: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    user,
    className,
    invitableCollaborators,
    isLoadingCollaborators,
}) => {
    const inviteEmailToSourcegraph = useInviteEmailToSourcegraph()
    const preventSubmit = useCallback((event: React.FormEvent<HTMLFormElement>): void => event.preventDefault(), [])
    const [query, setQuery] = useState('')
    const { setStep } = useSteps()

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

    const hasInviteableCollaborators = isLoadingCollaborators || invitableCollaborators.length > 0
    const hasFilteredCollaborators = isLoadingCollaborators || filteredCollaborators.length > 0

    const inviteURL = `${window.context.externalURL}/sign-up?invitedBy=${user.username}`
    return (
        <div className={classNames(className, 'mx-2')}>
            <div className={styles.titleDescription}>
                <Typography.H3>Introduce friends and colleagues to Sourcegraph</Typography.H3>
                <p className="text-muted mb-4">
                    We’ll look for a few collaborators you might want to invite to Sourcegraph. These users won’t be
                    able to see your code unless they have access to it on the code host and also add that code to
                    Sourcegraph.
                </p>
            </div>
            {isErrorLike(inviteError) && <ErrorAlert error={inviteError} />}
            <div className="border overflow-hidden rounded">
                <header>
                    <div className="py-3 px-3 d-flex justify-content-between align-items-center">
                        <Typography.H4 className="flex-1 m-0">Collaborators</Typography.H4>
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
                                disabled={!hasInviteableCollaborators}
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
                                <UserAvatar inline={true} className={classNames('mr-3', styles.avatar)} user={person} />
                                <div>
                                    <strong>{person.displayName}</strong>
                                    <div className="text-muted">{person.email}</div>
                                </div>
                                {loadingInvites.has(person.email) ? (
                                    <LoadingSpinner inline={true} className={classNames('ml-auto', 'mr-3')} />
                                ) : successfulInvites.has(person.email) ? (
                                    <span className="text-muted ml-auto mr-3">
                                        <Icon className="mr-1" as={CheckCircleIcon} />
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
                    {!hasInviteableCollaborators ? (
                        <div className={styles.noCollaborators}>
                            <div>No collaborators found in your selected repositories.</div>
                            <div className="text-muted mt-3">
                                You can{' '}
                                <Link
                                    to="/welcome"
                                    onClick={event => {
                                        event.preventDefault()
                                        event.stopPropagation()
                                        setStep(SELECT_REPOS_STEP)
                                    }}
                                >
                                    select additional repositories
                                </Link>{' '}
                                or invite people to Sourcegraph with the direct invite link below.
                            </div>
                        </div>
                    ) : !hasFilteredCollaborators ? (
                        <div className={styles.noCollaborators}>
                            <div className="d-flex text-muted justify-content-center mt-3">
                                No collaborators found with your search filter.
                            </div>
                        </div>
                    ) : null}
                </div>
                <div className={styles.inviteAllContainer}>
                    {hasInviteableCollaborators ? (
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
                    ) : null}
                </div>
            </div>
            <div>
                <header>
                    <div className="py-3 d-flex justify-content-between align-items-center">
                        <Typography.H4 className="m-0">Or invite by sending a link</Typography.H4>
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
