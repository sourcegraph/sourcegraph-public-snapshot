import { useMemo, type FC } from 'react'

import { Alert, Text, Input, useDebounce, Link } from '@sourcegraph/wildcard'

interface ExternalUrlFormProps {
    className?: string
    url?: string
    onChange: (newUrl: string) => void
}
export const ExternalUrlForm: FC<ExternalUrlFormProps> = ({ className, url = '', onChange }) => {
    const debouncedUrl = useDebounce(url, 500)

    const isSameUrl = useMemo(() => window.location.href.startsWith(debouncedUrl), [debouncedUrl])

    return (
        <div className={className}>
            <Text>
                Configure the URL your organization will use to access this Sourcegraph instance. An external URL is
                required in order for certain features on Sourcegraph to work correctly. See{' '}
                <Link to="/help/admin/url">documentation</Link> for more information.
            </Text>
            {!debouncedUrl && <Alert variant="danger">You have not yet configured an external URL.</Alert>}
            {!isSameUrl && (
                <Alert variant="warning">
                    The configured URL does not match the current URL. You may need to update your DNS records.
                </Alert>
            )}

            <Input
                placeholder="https://sourcegraph.example.com"
                onChange={event => onChange(event.target.value)}
                value={url}
                aria-label="External URL"
            />
        </div>
    )
}
