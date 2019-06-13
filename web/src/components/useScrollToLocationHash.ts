import H from 'history'
import { useEffect } from 'react'

/**
 * A React hook that scrolls the viewport to the element identified in the location hash (e.g., the
 * element with ID "foo" for the URL "https://example.com/a/b/#foo").
 *
 * This is needed because the browser's standard scroll-to-hash behavior doesn't work when using
 * react-router.
 */
export const useScrollToLocationHash = (
    location: H.Location,
    globalDocument: Document | { getElementById(id: string): { scrollIntoView(): void } | null } = document
): void => {
    useEffect(() => {
        if (location.hash) {
            const element = globalDocument.getElementById(location.hash.slice(1))
            if (element) {
                element.scrollIntoView()
            }
        }
    }, [globalDocument, location.hash])
}
