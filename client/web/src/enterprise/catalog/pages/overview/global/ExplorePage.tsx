import React, { useEffect, useRef, useState } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Page } from '@sourcegraph/web/src/components/Page'
import { Button, PageHeader } from '@sourcegraph/wildcard'

import { CatalogIcon } from '../../../../../catalog'
import { Badge } from '../../../../../components/Badge'
import { FeedbackPromptContent } from '../../../../../nav/Feedback/FeedbackPrompt'
import { Popover } from '../../../../insights/components/popover/Popover'
import { CatalogExplorer } from '../components/catalog-explorer/CatalogExplorer'

import styles from './ExplorePage.module.scss'

interface Props extends TelemetryProps {}

/**
 * The catalog overview page.
 */
export const ExplorePage: React.FunctionComponent<Props> = ({ telemetryService }) => {
    useEffect(() => {
        telemetryService.logViewEvent('CatalogExplore')
    }, [telemetryService])

    return (
        <Page>
            <PageHeader
                path={[{ icon: CatalogIcon, text: 'Catalog' }]}
                className="mb-4"
                description="Explore software components, services, libraries, APIs, and more."
                actions={<FeedbackPopoverButton />}
            />
            <CatalogExplorer />
        </Page>
    )
}

const FeedbackPopoverButton: React.FunctionComponent = () => {
    const buttonReference = useRef<HTMLButtonElement>(null)
    const [isVisible, setVisibility] = useState(false)

    return (
        <div className="d-flex align-items-center px-2">
            <Badge status="wip" className="text-uppercase mr-2" />
            <Button ref={buttonReference} variant="link" size="sm">
                Share feedback
            </Button>
            <Popover
                isOpen={isVisible}
                target={buttonReference}
                onVisibilityChange={setVisibility}
                className={styles.feedbackPrompt}
            >
                <FeedbackPromptContent
                    closePrompt={() => setVisibility(false)}
                    textPrefix="Catalog: "
                    routeMatch="/catalog"
                />
            </Popover>
        </div>
    )
}
