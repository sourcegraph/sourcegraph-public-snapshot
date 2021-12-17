import { gql, useMutation } from '@apollo/client'
import React, { useCallback, useState } from 'react'

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

export const TosConsentModal: React.FunctionComponent<{ afterTosAccepted: () => void }> = ({ afterTosAccepted }) => {
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
                <h1>We respect your data privacy</h1>
                <p className="mb-5">
                    We take data privacy seriously. We collect only what we need to provide a great experience, and we
                    never have access to your private data or code.
                </p>
                <form onSubmit={onSubmit}>
                    <div className="form-group">
                        <div className="form-check">
                            <label className="form-check-label">
                                <input type="checkbox" className="form-check-input" onChange={onAgreeChanged} /> I agree
                                to Sourcegraph's{' '}
                                <a href="https://about.sourcegraph.com/terms" target="_blank" rel="noopener">
                                    Terms of Service
                                </a>{' '}
                                and{' '}
                                <a href="https://about.sourcegraph.com/privacy" target="_blank" rel="noopener">
                                    Privacy Policy
                                </a>{' '}
                                (required)
                            </label>
                        </div>
                    </div>
                    <LoaderButton
                        loading={loading}
                        label="Agree and continue"
                        type="submit"
                        disabled={!agree}
                        className="btn btn-primary mt-4"
                    />
                </form>
                <p className="mt-5">
                    If you do not agree, <a href="/-/sign-out">Sign Out</a> and contact your site admin to have your
                    account deleted.
                </p>
                {error && (
                    <div className="alert alert-danger mt-4">Error accepting Terms of Service: {error.message}</div>
                )}
            </div>
        </div>
    )
}
