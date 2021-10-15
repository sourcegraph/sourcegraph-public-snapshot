// "Error code" constants for Sourcegraph URL validation
export const URL_FETCH_ERROR = 'URL_FETCH_ERROR'
export const URL_AUTH_ERROR = 'URL_AUTH_ERROR'
export const LINK_PROPS: Pick<React.AnchorHTMLAttributes<HTMLAnchorElement>, 'rel' | 'target'> = {
    target: '_blank',
    rel: 'noopener noreferrer',
}
