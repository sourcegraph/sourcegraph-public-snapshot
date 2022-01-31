import React, { useEffect } from 'react'

import { CtaAlert } from '@sourcegraph/shared/src/components/CtaAlert'

import { AtomIcon, JetBrainsIcon, SublimeTextIcon, VSCodeIcon } from '../../components/CtaIcons'
import { eventLogger } from '../../tracking/eventLogger'

interface Props {
    className: string
    onAlertDismissed: () => void
}

export const IDEExtensionAlert: React.FunctionComponent<Props> = ({ className, onAlertDismissed }) => {
    useEffect(() => {
        eventLogger.log('InstallIDEExtensionCTAShown')
    }, [])

    return (
        <CtaAlert
            title="The power of Sourcegraph in your IDE"
            description="Link your editor and Sourcegraph to improve your ability to reference and reuse code across multiple repositories."
            cta={{
                label: 'Learn more about IDE extensions',
                href:
                    'https://docs.sourcegraph.com/integration/editor?utm_medium=inproduct&utm_source=search&utm_campaign=inproduct-cta&utm_term=null',
                onClick: onIDEExtensionClick,
            }}
            icon={
                <>
                    <VSCodeIcon />
                    <JetBrainsIcon />
                    <SublimeTextIcon />
                    <AtomIcon />
                </>
            }
            className={className}
            onClose={onAlertDismissed}
        />
    )
}

const onIDEExtensionClick = (): void => {
    eventLogger.log('InstallIDEExtensionCTAClicked')
}
