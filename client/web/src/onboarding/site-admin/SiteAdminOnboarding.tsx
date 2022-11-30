import React, { useCallback, useEffect, useState, useMemo } from 'react'

import { mdiCheckboxBlankCircleOutline, mdiClose, mdiCheckboxMarkedCircle } from '@mdi/js'
import classNames from 'classnames'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Alert, Button, H2, H4, Icon, Link, Text } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../../backend/graphql'
import { SiteAdminOnboardingResult } from '../../graphql-operations'

import styles from './SiteAdminOnboarding.module.scss'

const SITE_ADMIN_ONBOARDING = gql`
    query SiteAdminOnboarding {
        # BEGIN Check SSO setup
        site {
            authProviders {
                nodes {
                    isBuiltin
                }
            }
        }
        # END Check SSO setup
        # BEGIN Check Code Host setup
        externalServices {
            totalCount
        }
        # END Check Code Host setup
        # BEGIN Check repositories setup
        repositories {
            totalCount
        }
        # END Check repositories setup
        users {
            totalCount
        }
        # BEGIN Check run first search
        currentUser {
            usageStatistics {
                searchQueries
                findReferencesActions
                codeIntelligenceActions
            }
        }
        # END Check run first search
    }
`

interface SiteAdminOnboardingStep {
    id: string
    title: string
    url?: string
    description: string
    isCompleted?: boolean
}

interface UseSiteAdminOnboardingReturnType {
    steps: SiteAdminOnboardingStep[]
    isLoading: boolean
}

function useSiteAdminOnboardingSteps(): UseSiteAdminOnboardingReturnType {
    const [steps, setSteps] = useState<SiteAdminOnboardingStep[]>([])
    const [isLoading, setIsLoading] = useState(true)
    useEffect((): void => {
        requestGraphQL<SiteAdminOnboardingResult>(SITE_ADMIN_ONBOARDING)
            .pipe(
                map(dataOrThrowErrors),
                map(
                    data =>
                        [
                            {
                                id: 'AddCodeHost',
                                title: 'Add your code host',
                                url: '/site-admin/external-services/new',
                                description:
                                    'Connect your code host to Sourcegraph to enable code intelligence and search across all your repositories.',
                                isCompleted: data.externalServices.totalCount > 0, // TODO: clarify whether this should track also that repositories has been added
                            },
                            // {
                            //     id: 'add-repositories',
                            //     title: 'Add repositories',
                            //     url: '/site-admin/external-services/new',
                            //     description:
                            //         'Add repositories to Sourcegraph to enable code intelligence and search across all your repositories.',
                            //     isCompleted: !!data.repositories?.totalCount && data.repositories?.totalCount > 0,
                            // },
                            {
                                id: 'ConfigureSSO',
                                title: 'Configure SSO',
                                url: '/help/admin/auth#user-authentication-sso',
                                description:
                                    'Configure a single-sign on (SSO) provider or have at least one other teammate sign up.',
                                isCompleted: data.site.authProviders.nodes.filter(node => !node.isBuiltin).length > 0,
                            },
                            {
                                id: 'RunSearch',
                                title: 'Search your code',
                                description: (
                                    <span>
                                        Perform a search query on your code. <strong>Example:</strong> type 'lang:' and
                                        select a language
                                    </span>
                                ),
                                // TODO: clarify whether this should be marked as completed if the user has performed a search without no repositories/code hosts
                                isCompleted:
                                    !!data.currentUser?.usageStatistics?.searchQueries &&
                                    data.currentUser?.usageStatistics?.searchQueries > 0,
                            },
                        ] as SiteAdminOnboardingStep[]
                )
            )
            .toPromise()
            .then(setSteps)
            .finally(() => setIsLoading(false))
            .catch(console.error)
    }, [])

    return {
        steps,
        isLoading,
    }
}

interface SiteAdminOnboardingContentProps extends TelemetryProps {
    className?: string
    onDismiss: () => void
}

const SiteAdminOnboardingContent: React.FunctionComponent<SiteAdminOnboardingContentProps> = ({
    className,
    onDismiss,
    telemetryService,
}) => {
    const { steps, isLoading } = useSiteAdminOnboardingSteps()

    const hasCompleted = useMemo(() => steps.every(step => step.isCompleted), [steps])

    if (isLoading) {
        return null
    }

    return (
        <Alert
            variant={hasCompleted ? 'success' : 'info'}
            className={classNames('py-4 pr-3 w-100 position-relative', className)}
        >
            <div className="mb-4 d-flex flex-column align-items-baseline">
                <H2 className="mr-2">Get Started</H2>
                <Text className="m-0">Quick steps to get started with Sourcegraph.</Text>
            </div>
            <div className={classNames('d-flex justify-content-between', styles.steps)}>
                {steps.map((step, index) => (
                    <div className="flex-grow-1" key={step.id}>
                        <H4>
                            {step.isCompleted ? (
                                <Icon
                                    svgPath={mdiCheckboxMarkedCircle}
                                    size="md"
                                    className="text-success"
                                    aria-label="Completed"
                                />
                            ) : (
                                <Icon
                                    svgPath={mdiCheckboxBlankCircleOutline}
                                    size="md"
                                    className="text-muted"
                                    aria-label="Close"
                                />
                            )}{' '}
                            {index + 1}.{' '}
                            {step.url ? (
                                <Link
                                    to={step.url}
                                    target="_blank"
                                    rel="noreferrer"
                                    onClick={() =>
                                        telemetryService.log(
                                            'SiteAdminOnboardingStepClicked',
                                            {
                                                id: step.id,
                                            },
                                            {
                                                id: step.id,
                                            }
                                        )
                                    }
                                >
                                    {step.title}
                                </Link>
                            ) : (
                                step.title
                            )}
                        </H4>
                        <Text className="m-0 flex-grow-1 d-flex align-items-center">{step.description}</Text>
                    </div>
                ))}
            </div>
            <Button onClick={onDismiss} size="sm" variant="icon" className={classNames(styles.dismissButton, 'p-2')}>
                <Icon svgPath={mdiClose} size="md" aria-label="Close" />
            </Button>
        </Alert>
    )
}

interface SiteAdminOnboardingProps extends Omit<SiteAdminOnboardingContentProps, 'onDismiss'> {}

export const SiteAdminOnboarding: React.FunctionComponent<SiteAdminOnboardingProps> = ({
    className,
    telemetryService,
}) => {
    const [hasDismissed, setHasDismissed] = useTemporarySetting('onboarding.admin.hasDismissed', false)

    const onDismiss = useCallback((): void => {
        setHasDismissed(true)
        telemetryService.log('SiteAdminOnboardingDismissed')
    }, [setHasDismissed, telemetryService])

    if (hasDismissed) {
        return null
    }

    return (
        <SiteAdminOnboardingContent className={className} onDismiss={onDismiss} telemetryService={telemetryService} />
    )
}
