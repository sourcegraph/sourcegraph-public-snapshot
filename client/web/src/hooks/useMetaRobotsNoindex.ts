import { useEffect } from 'react'

/**
 * Make a page non-indexable by search engines by passing `true`
 *
 * @param allow Whether indexing should be allowed or not
 */
export function useMetaRobotsNoIndex(allow: boolean): void {
    useEffect(() => {
        const metaTag = document.querySelector<HTMLMetaElement>('meta[name="robots"]')
        if (metaTag) {
            metaTag.content = allow ? '__META_ROBOTS_CONTENT__' : 'noindex, nofollow'
        }
    }, [allow])
}
