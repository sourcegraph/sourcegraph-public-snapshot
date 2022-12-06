import React from 'react'

import { mdiCodeBraces, mdiLock, mdiChevronRight } from '@mdi/js'
import classNames from 'classnames'

import { gql, useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { H3, Text, Link, Icon } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { MarketingBlock } from '../components/MarketingBlock'
import { useFeatureFlag } from '../featureFlags/useFeatureFlag'
import { ExternalServicesTotalCountResult } from '../graphql-operations'

import styles from './AddCodeHostWidget.module.scss'

const EXTERNAL_SERVICES_TOTAL_COUNT = gql`
    query ExternalServicesTotalCount {
        externalServices {
            totalCount
        }
    }
`

export function useShouldShowAddCodeHostWidget(authenticatedUser: AuthenticatedUser | null): boolean | undefined {
    const [isAddCodeHostWidgetEnabled] = useFeatureFlag('plg-enable-add-codehost-widget', false)
    const { data } = useQuery<ExternalServicesTotalCountResult>(EXTERNAL_SERVICES_TOTAL_COUNT, {})

    return (
        isAddCodeHostWidgetEnabled && authenticatedUser?.siteAdmin && !!data && data.externalServices.totalCount === 0
    )
}

interface AddCodeHostWidgetProps extends TelemetryProps {
    className?: string
}

export const AddCodeHostWidget: React.FunctionComponent<AddCodeHostWidgetProps> = ({ className, telemetryService }) => (
    <MarketingBlock
        wrapperClassName={classNames('mt-3 mx-auto', className, styles.container)}
        contentClassName={classNames(styles.innerContainer, 'p-4 d-flex flex-column align-items-baseline')}
    >
        <H3 className="mr-2 mb-1">Let’s start by adding your code</H3>
        <Text
            as={Link}
            to="/site-admin/external-services/new"
            className="d-inline-flex align-items-center"
            weight="medium"
            onClick={() => telemetryService.log('OnboardingWidget:ConnectCodeHost:Clicked')}
        >
            Connect code host
            <Icon svgPath={mdiChevronRight} className="ml-1" size="md" aria-label="Arrow right icon" />
        </Text>
        <div className={classNames(styles.divider, 'w-100 my-3')} />
        <div className="d-flex mb-2">
            <Icon svgPath={mdiCodeBraces} size="md" className="mr-2" aria-label="Code brace icon" />
            <div>
                <Text weight="bold" className="mb-1">
                    How does Sourcegraph connect to my code?
                </Text>
                <Text className="text-muted">
                    Sourcegraph syncs repositories from code hosts and other similar services (
                    <Link
                        to="/help/admin/external_service"
                        target="_blank"
                        rel="noopener noreferrer"
                        className={styles.textUnderline}
                        onClick={() => telemetryService.log('OnboardingWidget:ViewDocs:Clicked')}
                    >
                        docs
                    </Link>
                    ).
                </Text>
            </div>
        </div>
        <div className="d-flex">
            <Icon svgPath={mdiLock} size="md" className="mr-2" aria-label="Lock icon" />
            <div>
                <Text weight="bold" className="mb-1">
                    Want to know more about Security?
                </Text>
                <Text className="text-muted m-0">
                    Sourcegraph has successfully completed a{' '}
                    <Link
                        to="https://security.sourcegraph.com/?itemUid=7bfa66da-33ab-49de-8391-e329738a1ae9&source=title"
                        target="_blank"
                        rel="noopener noreferrer"
                        className={styles.textUnderline}
                        onClick={() => telemetryService.log('OnboardingWidget:ViewSOC2:Clicked')}
                    >
                        SOC 2 Type 2
                    </Link>{' '}
                    audit. Read about this and more on our{' '}
                    <Link
                        to="https://security.sourcegraph.com/"
                        target="_blank"
                        rel="noopener noreferrer"
                        className={styles.textUnderline}
                        onClick={() => telemetryService.log('OnboardingWidget:ViewSecurityPortal:Clicked')}
                    >
                        security portal
                    </Link>
                    .
                </Text>
            </div>
        </div>
    </MarketingBlock>
)
