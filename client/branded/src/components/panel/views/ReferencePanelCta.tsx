import React from 'react'

import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { Alert, Button } from '@sourcegraph/wildcard'

import styles from './ReferencePanelCta.module.scss'

export const ReferencePanelCta: React.FunctionComponent = () => {
    // Determine if we should show the CTA at all. The initial value will be
    // the current user's temporary setting (so we can show it until they interact).
    const [ctaDismissed, setCtaDismissed] = useTemporarySetting('codeintel.referencePanel.redesign.ctaDismissed', false)
    const [, setEnabled] = useTemporarySetting('codeintel.referencePanel.redesign.enabled', false)

    return (
        <>
            {ctaDismissed === false && (
                <Alert className={classNames('mr-4', styles.container)} variant="info">
                    <span>
                        Try out our{' '}
                        <Button
                            variant="link"
                            className={classNames('m-0 p-0 border-0', styles.button)}
                            onClick={() => setEnabled(true)}
                        >
                            brand new reference panel experience
                        </Button>
                    </span>
                    <Button
                        variant="link"
                        className={classNames('m-0 p-0 text-right', styles.button)}
                        onClick={() => setCtaDismissed(true)}
                    >
                        <CloseIcon size={16} />
                    </Button>
                </Alert>
            )}
        </>
    )
}
