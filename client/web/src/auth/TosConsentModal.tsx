import React, { useCallback, useState } from 'react'

import { gql, useMutation } from '@apollo/client'

import { Link, Alert, AnchorLink, Checkbox, Typography } from '@sourcegraph/wildcard'

import { LoaderButton } from '../components/LoaderButton'

import { SourcegraphIcon } from './icons'

import styles from './TosConsentModal.module.scss'

export const SET_TOS_ACCEPTED_MUTATION = gql`
    mutation {
        setTosAccepted {
            alwaysNil
        }
    }
`

export const TosConsentModal: React.FunctionComponent<React.PropsWithChildren<{ afterTosAccepted: () => void }>> = ({
    afterTosAccepted,
}) => {
    const [agree, setAgree] = useState(false)

    const onAgreeChanged = useCallback((event: React.ChangeEvent<HTMLInputElement>): void => {
        setAgree(event.target.checked)
    }, [])

    const [setTosAccepted, { loading, error }] = useMutation(SET_TOS_ACCEPTED_MUTATION)

    const onSubmit = useCallback(
        async (event: React.FormEvent<HTMLFormElement>): Promise<void> => {
            event.preventDefault()

            try {
                await setTosAccepted()
                afterTosAccepted()
            } catch (error) {
                console.error(error)
            }
        },
        [afterTosAccepted, setTosAccepted]
    )

    return (
        <div className={styles.container}>
            <SourcegraphIcon className={styles.icon} />
            <div className={styles.content}>
                <Typography.H1>We respect your data privacy</Typography.H1>
                <p className="mb-5">
                    We take data privacy seriously. We collect only what we need to provide a great experience, and we
                    never have access to your private data or code.
                </p>
                {/* eslint-disable-next-line react/forbid-elements */}
                <form onSubmit={onSubmit}>
                    <div className="form-group">
                        <Checkbox
                            onChange={onAgreeChanged}
                            id="terms-and-services-checkbox"
                            label={
                                <>
                                    I agree to Sourcegraph's{' '}
                                    <Link to="https://about.sourcegraph.com/terms" target="_blank" rel="noopener">
                                        Terms of Service
                                    </Link>{' '}
                                    and{' '}
                                    <Link to="https://about.sourcegraph.com/privacy" target="_blank" rel="noopener">
                                        Privacy Policy
                                    </Link>{' '}
                                    (required)
                                </>
                            }
                        />
                    </div>
                    <LoaderButton
                        loading={loading}
                        label="Agree and continue"
                        type="submit"
                        disabled={!agree}
                        className="mt-4"
                        variant="primary"
                    />
                </form>
                <p className="mt-5">
                    If you do not agree, <AnchorLink to="/-/sign-out">sign out</AnchorLink> and contact your site admin
                    to have your account deleted.
                </p>
                {error && (
                    <Alert className="mt-4" variant="danger">
                        Error accepting Terms of Service: {error.message}
                    </Alert>
                )}
            </div>
        </div>
    )
}
