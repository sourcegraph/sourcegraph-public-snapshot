import { useState } from 'react'

import { Alert, Button, Container, H3, LoadingSpinner, Text } from '@sourcegraph/wildcard'

import { UpdateInfo } from './updater'

interface UpdateInfoContentProps {
    details: UpdateInfo
}

function UpdateDetails({ details }: UpdateInfoContentProps): JSX.Element {
    const [showDetails, setShowDetails] = useState<boolean>(false)

    return (
        <div>
            <Text className="mb-0">
                Cody version upgrade from {details.version} to <b>{details.newVersion}</b> update available.
            </Text>
            <Button
                variant="link"
                onClick={() => setShowDetails(true)}
                disabled={showDetails || details.description === undefined}
            >
                Details
            </Button>
            |
            <Button
                variant="link"
                onClick={details.startInstall}
                disabled={details.stage !== 'IDLE' || details.startInstall === undefined}
            >
                Install Now
            </Button>
            {['INSTALLING', 'PENDING'].includes(details.stage) && (
                <Container className="d-flex p-2 mt-2">
                    <LoadingSpinner />
                    <Text className="ml-2 mb-1">Installing... Please wait...</Text>
                </Container>
            )}
            {details.stage === 'ERROR' && (
                <div className="mt-2">
                    <Alert variant="danger">
                        {details.error}
                        {details.checkNow !== undefined && (
                            <Button
                                variant="link"
                                onClick={() => {
                                    details.checkNow?.(true)
                                }}
                            >
                                Try Again
                            </Button>
                        )}
                    </Alert>
                </div>
            )}
            {showDetails && (
                <Container className="p-2 mt-2">
                    <H3>Version {details.newVersion} Details</H3>
                    <Text className="m-0">{details.description}</Text>
                </Container>
            )}
        </div>
    )
}

export function UpdateInfoContent({ details: update }: UpdateInfoContentProps): JSX.Element {
    return update.stage === 'CHECKING' ? (
        <div className="d-flex">
            <LoadingSpinner inline={true} />
            <Text className="ml-2 mb-1">Please wait... Checking for updates...</Text>
        </div>
    ) : update.hasNewVersion ? (
        <UpdateDetails details={update} />
    ) : (
        <Text className="mb-1">
            No updates needed. Application already up to date at {update.version}.
            {update.checkNow !== undefined && (
                <Button
                    variant="link"
                    onClick={() => {
                        update.checkNow?.(true)
                    }}
                >
                    Check Now
                </Button>
            )}
        </Text>
    )
}
