import React from 'react'

import { Button, Link } from '@sourcegraph/wildcard'

import { Toast } from './Toast'

import styles from './Step.module.scss'

interface thankYouProps {
    handleSubmit: () => void
}

export const ThankYou: React.FunctionComponent<thankYouProps> = ({ handleSubmit }) => {
    const LINKS = [
        { label: 'How Ebay uses search notebooks to onboard new teammates', link: 'ebay.com' },
        { label: 'How Ebay uses search notebooks to onboard new teammates', link: 'ebay.com' },
        { label: 'How Ebay uses search notebooks to onboard new teammates', link: 'ebay.com' },
    ]

    return (
        <Toast
            className={styles.toast}
            subtitle={<span className={styles.toastSubtitle}>Thank you for your feedback</span>}
            cta={
                <div className={styles.thankyouToastContent}>
                    <div className="mb-4">
                        You can learn more about using Sourcegraph to onboard to a new code base with these resources{' '}
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
                <div className={styles.done}>
                    <Button id="survey-toast-dismiss" variant="primary" size="sm" onClick={handleSubmit}>
                        Done
                    </Button>
                </div>
            }
            onDismiss={handleSubmit}
        />
    )
}
