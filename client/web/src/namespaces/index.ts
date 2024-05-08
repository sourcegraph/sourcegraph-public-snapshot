import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'

import type { OrgAreaOrganizationFields, UserAreaUserFields } from '../graphql-operations'

/**
 * Common props for components underneath a namespace (e.g., a user or organization).
 */
export interface NamespaceProps extends TelemetryV2Props {
    /**
     * The namespace.
     */
    namespace: Pick<UserAreaUserFields | OrgAreaOrganizationFields, '__typename' | 'id' | 'url'>
}
