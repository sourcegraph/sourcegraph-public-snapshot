export interface UTMMarker {
    utm_source?: string
    utm_campaign?: string
    utm_medium?: string
    utm_term?: string
    utm_content?: string
}

/**
 * Returns a new URL with UTM markers appended as query parameters
 *
 * @param url URL to which append UTM markers
 * @param utm UTM markers to append
 */
export const createURLWithUTM = (url: URL, utm: UTMMarker): URL => {
    const result = new URL(url)
    for (const [key, value] of Object.entries(utm)) {
        result.searchParams.set(key, value)
    }
    return result
}
