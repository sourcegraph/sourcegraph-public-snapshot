import { type FC, useMemo } from 'react'

import { mdiInformation } from '@mdi/js'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import GitIcon from 'mdi-react/GitIcon'
import GitLabIcon from 'mdi-react/GitlabIcon'
import { useLocation } from 'react-router-dom'

import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Link, Alert, H3, Icon, Text, Container, PageHeader } from '@sourcegraph/wildcard'

import { ChecklistInfo } from '../../site-admin/setup-checklist/ChecklistInfo'
import { LimitedAccessBanner } from '../LimitedAccessBanner'
import { PageTitle } from '../PageTitle'

import { AddExternalServicePage } from './AddExternalServicePage'
import { ExternalServiceGroup, AddExternalServiceOptionsWithID } from './ExternalServiceGroup'
import { allExternalServices, AddExternalServiceOptions, gitHubAppConfig } from './externalServices'

export interface AddExternalServicesPageProps extends TelemetryProps {
    /**
     * The list of code host external services to be displayed.
     * Pick items from externalServices.codeHostExternalServices.
     */
    codeHostExternalServices: Record<string, AddExternalServiceOptions>
    /**
     * The list of non-code host external services to be displayed.
     * Pick items from externalServices.nonCodeHostExternalServices.
     */
    nonCodeHostExternalServices: Record<string, AddExternalServiceOptions>

    externalServicesFromFile: boolean
    allowEditExternalServicesWithFile: boolean
    isSourcegraphApp: boolean

    /** For testing only. */
    autoFocusForm?: boolean
}

/**
 * Page for choosing a service kind and variant to add, among the available options.
 */
export const AddExternalServicesPage: FC<AddExternalServicesPageProps> = ({
    codeHostExternalServices,
    nonCodeHostExternalServices,
    telemetryService,
    autoFocusForm,
    externalServicesFromFile,
    allowEditExternalServicesWithFile,
    isSourcegraphApp,
}) => {
    const { search } = useLocation()
    const [hasDismissedPrivacyWarning, setHasDismissedPrivacyWarning] = useTemporarySetting(
        'admin.hasDismissedCodeHostPrivacyWarning',
        false
    )
    const dismissPrivacyWarning = (): void => {
        setHasDismissedPrivacyWarning(true)
    }

    const externalService = useMemo(() => {
        const params = new URLSearchParams(search)
        const id = params.get('id')
        if (id) {
            let externalService = allExternalServices[id]
            if (externalService?.kind === ExternalServiceKind.GITHUB) {
                const appID = params.get('appID')
                const installationID = params.get('installationID')
                const baseURL = params.get('url')
                if (externalService === codeHostExternalServices.ghapp) {
                    externalService = gitHubAppConfig(baseURL, appID, installationID)
                }
            }
            return externalService
        }
        return null
    }, [search, codeHostExternalServices.ghapp])

    if (externalService) {
        return (
            <AddExternalServicePage
                telemetryService={telemetryService}
                externalService={externalService}
                autoFocusForm={autoFocusForm}
                externalServicesFromFile={externalServicesFromFile}
                allowEditExternalServicesWithFile={allowEditExternalServicesWithFile}
            />
        )
    }

    const licenseInfo = window.context.licenseInfo
    let allowedCodeHosts: AddExternalServiceOptions[] | null = null
    // if (licenseInfo && licenseInfo.currentPlan === 'business-0') {
    if (!(licenseInfo && licenseInfo.currentPlan === 'business-0')) {
        allowedCodeHosts = [
            codeHostExternalServices.github,
            codeHostExternalServices.gitlabcom,
            codeHostExternalServices.bitbucket,
        ]
    }

    const codeHostServicesGroup = computeExternalServicesGroup(codeHostExternalServices, allowedCodeHosts)

    return (
        <>
            <PageTitle title="Add a code host connection" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Add a code host connection' }]}
                description="Add code host connection to one of the supported code hosts."
                className="mb-3"
            />

            {isSourcegraphApp && (
                <LimitedAccessBanner
                    storageKey="app.manage-repositories-with-new-settings"
                    badgeText="Repositories"
                    className="mb-3"
                >
                    Manage your local repositories in your settings. Go to{' '}
                    <Link to="/user/app-settings">Settings → Repositories → Local/Remote repositories</Link>
                </LimitedAccessBanner>
            )}

            <Container className="mb-3">
                <ChecklistInfo />
                {!hasDismissedPrivacyWarning && (
                    <ExternalServicesPrivacyAlert dismissPrivacyWarning={dismissPrivacyWarning} />
                )}

                {allowedCodeHosts && (
                    <>
                        <Text>
                            <Icon aria-label="Information icon" svgPath={mdiInformation} /> Upgrade to{' '}
                            <Link to="https://about.sourcegraph.com/pricing">Sourcegraph Enterprise</Link> to add
                            repositories from other code hosts.
                        </Text>
                    </>
                )}

                {Object.entries(codeHostServicesGroup).map(([displayName, info], index) => (
                    <ExternalServiceGroup
                        // We ignore the index key rule here since the grouping doesn't have a
                        // unique identifier.
                        //
                        // eslint-disable-next-line react/no-array-index-key
                        key={`${index}-${displayName}`}
                        name={displayName}
                        services={info.services}
                        description={info.description}
                        icon={info.icon}
                        renderServiceIcon={info.renderServiceIcon}
                    />
                ))}

                {Object.values(nonCodeHostExternalServices).length > 0 && (
                    <ExternalServiceGroup
                        name="Dependencies"
                        services={transformExternalServices(nonCodeHostExternalServices)}
                        description=""
                        renderServiceIcon={true}
                    />
                )}
            </Container>
        </>
    )
}

interface ExternalServicesPrivacyAlertProps {
    dismissPrivacyWarning: () => void
}

const ExternalServicesPrivacyAlert: FC<ExternalServicesPrivacyAlertProps> = ({ dismissPrivacyWarning }) => (
    <Alert variant="info">
        <Text>
            This Sourcegraph installation will never send your code, repository names, file names, or any other specific
            code data to Sourcegraph.com or any other destination. Your code is kept private on this installation.
        </Text>
        <H3>This Sourcegraph installation will access your code host by:</H3>
        <ul>
            <li>
                Periodically fetching a list of repositories to ensure new, removed, and renamed repositories are
                accessible on Sourcegraph.
            </li>
            <li>Cloning the repositories you specify to create a local cache.</li>
            <li>Periodically pulling cloned repositories to ensure search results are current.</li>
            <li>
                Fetching{' '}
                <Link to="/help/admin/permissions" target="_blank" rel="noopener noreferrer">
                    user repository access permissions
                </Link>
                , if you have enabled this feature.
            </li>
            <li>
                Opening pull requests and syncing their metadata as part of{' '}
                <Link to="/help/batch_changes" target="_blank" rel="noopener noreferrer">
                    batch changes
                </Link>
                , if you have enabled this feature.
            </li>
        </ul>
        <div className="d-flex justify-content-end">
            <Button variant="secondary" onClick={dismissPrivacyWarning}>
                Do not show this again
            </Button>
        </div>
    </Alert>
)

interface ExternalServicesGroup {
    services: AddExternalServiceOptionsWithID[]
    icon: React.ComponentType<{ className?: string }>
    description: string
    renderServiceIcon: boolean
}

type GroupedServiceDisplayName = 'GitHub' | 'GitLab' | 'Bitbucket' | 'Other code hosts'

const computeExternalServicesGroup = (
    services: Record<string, AddExternalServiceOptions>,
    allowedCodeHosts: AddExternalServiceOptions[] | null
): Record<GroupedServiceDisplayName, ExternalServicesGroup> => {
    const groupedServices: Record<GroupedServiceDisplayName, ExternalServicesGroup> = {
        GitHub: {
            services: [],
            icon: GithubIcon,
            description: 'Connect with repositories on GitHub',
            renderServiceIcon: false,
        },
        GitLab: {
            services: [],
            icon: GitLabIcon,
            description: 'Connect with repositories on GitLab',
            renderServiceIcon: false,
        },
        Bitbucket: {
            services: [],
            icon: BitbucketIcon,
            description: 'Connect with repositories on Bitbucket',
            renderServiceIcon: false,
        },
        'Other code hosts': { services: [], icon: GitIcon, description: '', renderServiceIcon: true },
    }

    for (const [serviceID, service] of Object.entries(services)) {
        const isDisabled = Boolean(allowedCodeHosts && !allowedCodeHosts.includes(service))
        const otherProps = isDisabled
            ? {
                  badge: 'enterprise',
                  tooltip: 'Upgrade to Sourcegraph Enterprise to add repositories from other code hosts',
              }
            : {}
        switch (service.kind) {
            case ExternalServiceKind.GITHUB:
                groupedServices.GitHub.services.push({ ...service, serviceID, enabled: !isDisabled, ...otherProps })
                break
            case ExternalServiceKind.GITLAB:
                groupedServices.GitLab.services.push({ ...service, serviceID })
                break
            case ExternalServiceKind.BITBUCKETCLOUD:
            case ExternalServiceKind.BITBUCKETSERVER:
                groupedServices.Bitbucket.services.push({ ...service, serviceID })
                break
            default:
                groupedServices['Other code hosts'].services.push({ ...service, serviceID })
        }
    }

    return groupedServices
}

const transformExternalServices = (
    services: Record<string, AddExternalServiceOptions>
): AddExternalServiceOptionsWithID[] =>
    Object.entries(services).map(([serviceID, service]) => ({ ...service, serviceID }))
