import React, { useState, FunctionComponent, useCallback } from 'react'

import classNames from 'classnames'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { gql, dataOrThrowErrors } from '@sourcegraph/http-client'
import { Select } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../../../backend/graphql'
import { LoaderButton } from '../../../components/LoaderButton'
import { SetUserEmailPrimaryResult, SetUserEmailPrimaryVariables, UserEmailsResult } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'

import styles from './SetUserPrimaryEmailForm.module.scss'

type UserEmail = (NonNullable<UserEmailsResult['node']> & { __typename: 'User' })['emails'][number]

interface Props {
    user: string
    emails: UserEmail[]
    onDidSet: () => void

    className?: string
}

type Status = undefined | 'loading' | ErrorLike

const findPrimaryEmail = (emails: UserEmail[]): string | undefined => emails.find(email => email.isPrimary)?.email

export const SetUserPrimaryEmailForm: FunctionComponent<React.PropsWithChildren<Props>> = ({
    user,
    emails,
    onDidSet,
    className,
}) => {
    const [primaryEmail, setPrimaryEmail] = useState<string | undefined>(findPrimaryEmail(emails))
    const [statusOrError, setStatusOrError] = useState<Status>()

    // options should include all verified emails + a primary one
    const options = emails.filter(email => email.verified || email.isPrimary).map(email => email.email)

    const onPrimaryEmailSelect: React.ChangeEventHandler<HTMLSelectElement> = event =>
        setPrimaryEmail(event.target.value)

    const onSubmit: React.FormEventHandler<HTMLFormElement> = useCallback(
        async event => {
            event.preventDefault()
            if (primaryEmail === undefined) {
                return
            }
            setStatusOrError('loading')

            try {
                dataOrThrowErrors(
                    await requestGraphQL<SetUserEmailPrimaryResult, SetUserEmailPrimaryVariables>(
                        gql`
                            mutation SetUserEmailPrimary($user: ID!, $email: String!) {
                                setUserEmailPrimary(user: $user, email: $email) {
                                    alwaysNil
                                }
                            }
                        `,
                        { user, email: primaryEmail }
                    ).toPromise()
                )

                eventLogger.log('UserEmailAddressSetAsPrimary')
                setStatusOrError(undefined)

                if (onDidSet) {
                    onDidSet()
                }
            } catch (error) {
                setStatusOrError(asError(error))
            }
        },
        [user, primaryEmail, onDidSet]
    )

    return (
        <div className={classNames('add-user-email-form', className)}>
            <Form onSubmit={onSubmit}>
                <div className={styles.formLine}>
                    <div className={styles.formSelect}>
                        <Select
                            label="Primary email address"
                            labelVariant="block"
                            id="setUserPrimaryEmailForm-email"
                            isCustomStyle={true}
                            selectClassName="mr-sm-2"
                            value={primaryEmail}
                            onChange={onPrimaryEmailSelect}
                            required={true}
                            disabled={options.length === 1 || statusOrError === 'loading'}
                        >
                            {options.map(email => (
                                <option key={email} value={email}>
                                    {email}
                                </option>
                            ))}
                        </Select>
                    </div>
                    <div className={styles.formButton}>
                        <LoaderButton
                            loading={statusOrError === 'loading'}
                            label="Save"
                            type="submit"
                            disabled={options.length === 1 || statusOrError === 'loading'}
                            variant="primary"
                        />
                    </div>
                </div>
            </Form>
            {isErrorLike(statusOrError) && <ErrorAlert className="mt-2" error={statusOrError} />}
        </div>
    )
}
