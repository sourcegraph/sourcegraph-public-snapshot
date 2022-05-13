import React from 'react'

import classNames from 'classnames'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { Alert, Button, Link } from '@sourcegraph/wildcard'

import styles from './ReferencesPanelFeedbackCta.module.scss'

export const ReferencesPanelFeedbackCta: React.FunctionComponent = () => {
    // Determine if we should show the CTA at all. The initial value will be
    // the current user's temporary setting (so we can show it until they interact).
    const [ctaDismissed, setCtaDismissed] = useTemporarySetting('codeintel.referencePanelFeedback.ctaDismissed', false)

    return (
        <>
            {ctaDismissed === false && (
                <Alert className={classNames('m-2 mr-3 ml-3', styles.container)} variant="info">
                    <span>
                        Please leave{' '}
                        <Link to="https://github.com/sourcegraph/sourcegraph/issues">
                            feedback for our reference panel!
                        </Link>
                    </span>
                    <Button
                        variant="link"
                        className={classNames('m-0 p-0 text-right', styles.button)}
                        onClick={() => setCtaDismissed(true)}
                    >
                        Dismiss
                    </Button>
                </Alert>
            )}
        </>
    )
}
