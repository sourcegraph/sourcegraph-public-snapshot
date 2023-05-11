import { mdiCog } from '@mdi/js'

import { Icon } from '@sourcegraph/wildcard'
import { BreadcrumbItem } from '@sourcegraph/wildcard/src/components/PageHeader'

import { GitHubAppByAppIDResult } from '../../graphql-operations'

import { ExternalServiceFieldsWithConfig } from './backend'
import { resolveExternalServiceCategory } from './externalServices'

export const getBreadCrumbs = (
    externalService?: ExternalServiceFieldsWithConfig,
    ghAppData?: GitHubAppByAppIDResult,
    isEdit: boolean = false
): BreadcrumbItem[] => {
    if (externalService) {
        const externalServiceCategory = resolveExternalServiceCategory(externalService)
        if (ghAppData?.gitHubAppByAppID?.id) {
            return [
                { to: '/site-admin/github-apps', text: 'GitHub Apps' },
                {
                    to: `/site-admin/github-apps/${ghAppData?.gitHubAppByAppID?.id}`,
                    text: ghAppData?.gitHubAppByAppID?.name,
                },
                {
                    text: (
                        <>
                            {externalServiceCategory && (
                                <Icon
                                    inline={true}
                                    as={externalServiceCategory.icon}
                                    aria-label="Code host logo"
                                    className="mr-2"
                                />
                            )}
                            {externalService.displayName}
                        </>
                    ),
                },
            ]
        }

        return [
            { icon: mdiCog },
            { to: '/site-admin/external-services', text: 'Code host connections' },
            {
                to: isEdit ? `/site-admin/external-services/${externalService.id}/` : undefined,
                text: (
                    <>
                        {externalServiceCategory && (
                            <Icon
                                inline={true}
                                as={externalServiceCategory.icon}
                                aria-label="Code host logo"
                                className="mr-2"
                            />
                        )}
                        {externalService.displayName}
                    </>
                ),
            },
        ]
    }

    return [{ icon: mdiCog }, { to: '/site-admin/external-services', text: 'Code host connections' }]
}
