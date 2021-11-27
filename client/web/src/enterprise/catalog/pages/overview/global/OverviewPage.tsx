import PlusIcon from 'mdi-react/PlusIcon'
import React, { useEffect, useRef, useState } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, PageHeader } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../../auth'
import { CatalogIcon } from '../../../../../catalog'
import { Badge } from '../../../../../components/Badge'
import { Page } from '../../../../../components/Page'
import { FeedbackPromptContent } from '../../../../../nav/Feedback/FeedbackPrompt'
import { flipRightPosition } from '../../../../insights/components/context-menu/utils'
import { Popover } from '../../../../insights/components/popover/Popover'
import { OverviewContent } from '../components/overview-content/OverviewContent'

import styles from './OverviewPage.module.scss'

// TODO(sqs): extract the Insights components used above to the shared components area

export interface OverviewPageProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser
}

/**
 * The catalog overview page.
 */
export const OverviewPage: React.FunctionComponent<OverviewPageProps> = props => {
    const { telemetryService } = props

    useEffect(() => {
        telemetryService.logViewEvent('CatalogOverview')
    }, [telemetryService])

    const handleAddCatalogDataClick = (): void => {
        telemetryService.log('CatalogAddDataClick')
    }

    return (
        <div className="w-100">
            <Page>
                <PageHeader
                    annotation={<PageAnnotation />}
                    path={[{ icon: CatalogIcon, text: 'Catalog' }]}
                    actions={
                        <>
                            <Link
                                to="TODO(sqs)"
                                className="btn btn-outline-secondary mr-2"
                                onClick={handleAddCatalogDataClick}
                            >
                                <PlusIcon className="icon-inline" /> Add catalog data
                            </Link>
                        </>
                    }
                    className="mb-3"
                />

                <OverviewContent telemetryService={telemetryService} />
            </Page>
        </div>
    )
}

const PageAnnotation: React.FunctionComponent = () => {
    const buttonReference = useRef<HTMLButtonElement>(null)
    const [isVisible, setVisibility] = useState(false)

    return (
        <div className="d-flex align-items-center">
            <a href="TODO(sqs)" target="_blank" rel="noopener">
                <Badge status="wip" className="text-uppercase" />
            </a>

            <Button ref={buttonReference} variant="link" size="sm">
                Share feedback
            </Button>

            <Popover
                isOpen={isVisible}
                target={buttonReference}
                position={flipRightPosition}
                onVisibilityChange={setVisibility}
                className={styles.feedbackPrompt}
            >
                <FeedbackPromptContent
                    closePrompt={() => setVisibility(false)}
                    textPrefix="Code Insights: "
                    routeMatch="/insights/dashboards"
                />
            </Popover>
        </div>
    )
}
