import type { LineOrPositionOrRange } from '$lib/common'
import { ExternalServiceKind, type ExternalLink } from '$lib/graphql-types'

interface GetExternalURLOptions {
    externalLink: Pick<ExternalLink, 'url' | 'serviceKind'>
    lineOrPosition?: LineOrPositionOrRange
}

/**
 * Extends the external link with additional information depending on the service kind.
 *
 * @param externalLink The external link to process.
 * @param lineOrPosition The line or position to add to the external link.
 * @returns The external URL.
 */
export function getExternalURL({ externalLink, lineOrPosition }: GetExternalURLOptions): string {
    let url = externalLink.url

    switch (externalLink.serviceKind) {
        case ExternalServiceKind.GITHUB: {
            // Add range or position path to the code host URL.
            if (lineOrPosition?.line !== undefined) {
                url += `#L${lineOrPosition.line}`

                if (lineOrPosition.endLine) {
                    url += `-L${lineOrPosition.endLine}`
                }
            }
            break
        }
        case ExternalServiceKind.GITLAB: {
            // Add range or position path to the code host URL.
            if (lineOrPosition?.line !== undefined) {
                url += `#L${lineOrPosition.line}`

                if (lineOrPosition.endLine) {
                    url += `-${lineOrPosition.endLine}`
                }
            }
            break
        }
    }

    return url
}
