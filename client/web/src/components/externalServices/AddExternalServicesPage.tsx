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

import { LimitedAccessBanner } from '../LimitedAccessBanner'
import { PageTitle } from '../PageTitle'

import { AddExternalServicePage } from './AddExternalServicePage'
import { ExternalServiceGroup, type AddExternalServiceOptionsWithID } from './ExternalServiceGroup'
import { allExternalServices, type AddExternalServiceOptions, gitHubAppConfig } from './externalServices'

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
    isCodyApp: boolean

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
    isCodyApp,
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

    const allowedCodeHosts = useMemo(
        () => computeAllowedCodehosts(codeHostExternalServices),
        [codeHostExternalServices]
    )
    const servicesByGroup = useMemo(
        () => computeExternalServicesGroup(codeHostExternalServices, allowedCodeHosts),
        [codeHostExternalServices, allowedCodeHosts]
    )

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

    return (
        <>
            <PageTitle title="Add a code host connection" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Add a code host connection' }]}
                description="Add code host connection to one of the supported code hosts."
                className="mb-3"
            />

            {isCodyApp && (
                <LimitedAccessBanner
                    storageKey="app.manage-repositories-with-new-settings"
                    badgeText="Repositories"
                    className="mb-3"
                >
                    Manage your local repositories in your settings. Go to{' '}
                    <Link to="/user/app-settings">Settings → Repositories → Local/Remote repositories</Link>
                </LimitedAccessBanner>
            )}

            <Container>
                {!hasDismissedPrivacyWarning && (
                    <ExternalServicesPrivacyAlert dismissPrivacyWarning={dismissPrivacyWarning} />
                )}

                {allowedCodeHosts && (
                    <>
                        <Text>
                            <Icon aria-label="Information icon" svgPath={mdiInformation} /> Upgrade to{' '}
                            <Link to="https://sourcegraph.com/pricing">Sourcegraph Enterprise</Link> to add
                            repositories from other code hosts.
                        </Text>
                    </>
                )}

                {Object.values(servicesByGroup).map((serviceInfo, index) => (
                    <ExternalServiceGroup
                        // We ignore the index key rule here since the grouping doesn't have a
                        // unique identifier.
                        //
                        // eslint-disable-next-line react/no-array-index-key
                        key={`${index}-${serviceInfo.label}`}
                        name={serviceInfo.label}
                        services={serviceInfo.services}
                        description={serviceInfo.description}
                        icon={serviceInfo.icon}
                        renderIcon={serviceInfo.renderIcon}
                    />
                ))}

                {Object.values(nonCodeHostExternalServices).length > 0 && (
                    <ExternalServiceGroup
                        name="Dependencies"
                        services={transformNonCodeHostExternalServices(nonCodeHostExternalServices, allowedCodeHosts)}
                        renderIcon={true}
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
        <Text>
            When <Link to="/help/cody/overview">Cody</Link> is enabled, some of your data, including code, repository
            names, file names, and other specific code details, might be shared temporarily with our trusted LLM
            partner. We have established a zero retention policy agreement with the LLM company, to ensure the highest
            levels of data protection and confidentiality.
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
    label: string
    services: AddExternalServiceOptionsWithID[]
    icon: React.ComponentType<{ className?: string }>
    description: string
    renderIcon: boolean
    renderAsSingleItem?: boolean
}

const computeExternalServicesGroup = (
    services: Record<string, AddExternalServiceOptions>,
    allowedCodeHosts: Set<AddExternalServiceOptions> | null
): Record<string, ExternalServicesGroup> => {
    const groupedServices: Record<string, ExternalServicesGroup> = {
        github: {
            label: 'GitHub',
            services: [],
            icon: GithubIcon,
            description: 'Connect with repositories on GitHub',
            renderIcon: false,
        },
        gitlab: {
            label: 'GitLab',
            services: [],
            icon: GitLabIcon,
            description: 'Connect with repositories on GitLab',
            renderIcon: false,
            renderAsSingleItem: true,
        },
        bitbucket: {
            label: 'Bitbucket',
            services: [],
            icon: BitbucketIcon,
            description: 'Connect with repositories on Bitbucket',
            renderIcon: false,
        },
        other: { label: 'Other code hosts', services: [], icon: GitIcon, description: '', renderIcon: true },
    }

    for (const [serviceID, service] of Object.entries(services)) {
        const isDisabled = allowedCodeHosts?.has(service)
        const otherProps = isDisabled
            ? {
                  badge: 'enterprise',
                  tooltip: 'Upgrade your license to Sourcegraph Enterprise to add repositories from other code hosts',
              }
            : {}
        let key = 'other'
        if (service.kind === ExternalServiceKind.GITHUB || service.kind === ExternalServiceKind.GITLAB) {
            key = service.kind.toLowerCase()
        } else if (
            service.kind === ExternalServiceKind.BITBUCKETCLOUD ||
            service.kind === ExternalServiceKind.BITBUCKETSERVER
        ) {
            key = 'bitbucket'
        }
        groupedServices[key].services.push({ ...service, serviceID, enabled: !isDisabled, ...otherProps })
    }

    return groupedServices
}

const transformNonCodeHostExternalServices = (
    services: Record<string, AddExternalServiceOptions>,
    allowedCodeHosts: Set<AddExternalServiceOptions> | null
): AddExternalServiceOptionsWithID[] =>
    Object.entries(services).map(([serviceID, service]) => {
        const isDisabled = allowedCodeHosts?.has(service)
        return { ...service, serviceID, enabled: !isDisabled }
    })

const computeAllowedCodehosts = (
    codeHostExternalServices: Record<string, AddExternalServiceOptions>
): Set<AddExternalServiceOptions> | null => {
    const licenseInfo = window.context.licenseInfo
    let allowedCodeHosts: Set<AddExternalServiceOptions> | null = null
    if (licenseInfo && licenseInfo.currentPlan === 'business-0') {
        allowedCodeHosts = new Set([
            codeHostExternalServices.github,
            codeHostExternalServices.gitlabcom,
            codeHostExternalServices.bitbucket,
        ])
    }
    return allowedCodeHosts
}
