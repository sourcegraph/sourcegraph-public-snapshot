import { mdiCog } from '@mdi/js'

import { Icon } from '@sourcegraph/wildcard'
// eslint-disable-next-line no-restricted-imports
import type { BreadcrumbItem } from '@sourcegraph/wildcard/src/components/PageHeader'

import type { ExternalServiceFieldsWithConfig } from './backend'
import { resolveExternalServiceCategory } from './externalServices'

export const getBreadCrumbs = (
    externalService?: ExternalServiceFieldsWithConfig,
    isEdit: boolean = false
): BreadcrumbItem[] => {
    if (!externalService) {
        return [{ icon: mdiCog }, { to: '/site-admin/external-services', text: 'Code host connections' }]
    }
    const externalServiceCategory = resolveExternalServiceCategory(externalService)

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
