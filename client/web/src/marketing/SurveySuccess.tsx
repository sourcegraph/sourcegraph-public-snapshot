import React from 'react'

import { Button, Link } from '@sourcegraph/wildcard'

import { Toast } from './Toast'

import styles from './SurveySuccess.module.scss'

interface SurveySuccessProps {
    handleDismiss: () => void
}

const LINKS = [
    { label: 'How Ebay uses search notebooks to onboard new teammates', link: 'ebay.com' },
    { label: 'Embedding search notebooks in documentation, for the win', link: 'ebay.com' },
    { label: 'How to link to multiple lines of code', link: 'ebay.com' },
]

export const SurveySuccess: React.FunctionComponent<SurveySuccessProps> = ({ handleDismiss }) => (
    <Toast
        subtitle={<span className={styles.toastSubtitle}>Thank you for your feedback!</span>}
        cta={
            <div className={styles.thankyouToastContent}>
                <div className="mb-4">
                    You can learn more about using Sourcegraph to onboard to a new code base with these resources:{' '}
                </div>
                <ul className={styles.refLinkList}>
                    {LINKS.map(link => (
                        <li key={link.link}>
                            <Link to={link.link}>{link.label}</Link>
                        </li>
                    ))}
                </ul>
            </div>
        }
        footer={
            <div className="d-flex justify-content-end">
                <Button variant="primary" size="sm" onClick={handleDismiss}>
                    Done
                </Button>
            </div>
        }
        onDismiss={handleDismiss}
    />
)
