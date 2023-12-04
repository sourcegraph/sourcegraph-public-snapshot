import { Alert, Link, Text } from '@sourcegraph/wildcard'

export const ScimAlert = (): JSX.Element => (
    <Alert className="mb-4" variant="info">
        <Text className="mb-0">
            This profile is managed by the organization's identity provider through SCIM. Some fields are disabled.{' '}
            <Link to="/help/admin/scim" className="text-nowrap">
                Learn more about SCIM
            </Link>
        </Text>
    </Alert>
)
