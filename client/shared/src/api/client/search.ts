// For search-related extension API features, such as query transformers

import { Observable, of } from 'rxjs'

import { SharedEventLogger } from '../sharedEventLogger'

/**
 * Executes search query transformers contributed by Sourcegraph extensions.
 */
export function transformSearchQuery({
    query,
    enableGoImportsSearchQueryTransform,
    eventLogger,
}: {
    query: string
    enableGoImportsSearchQueryTransform: undefined | boolean
    eventLogger: SharedEventLogger
}): Observable<string> {
    // We apply any non-extension transform before we send the query to the
    // extensions since we want these to take presedence over the extensions.
    if (enableGoImportsSearchQueryTransform === undefined || enableGoImportsSearchQueryTransform) {
        query = goImportsTransform(query, eventLogger)
    }

    return of(query)
}

function goImportsTransform(query: string, eventLogger: SharedEventLogger): string {
    const goImportsRegex = /\bgo.imports:(\S*)/
    if (query.match(goImportsRegex)) {
        // Get package name
        const packageFilter = query.match(goImportsRegex)
        const packageName = packageFilter && packageFilter.length >= 1 ? packageFilter[1] : ''

        // Package imported in grouped import statements
        const matchPackage = '^\\t"[^\\s]*' + packageName + '[^\\s]*"$'
        // Match packages with aliases
        const matchAlias = '\\t[\\w/]*\\s"[^\\s]*' + packageName + '[^\\s]*"$'
        // Match packages in single import statement
        const matchSingle = 'import\\s"[^\\s]*' + packageName + '[^\\s]*"$'
        const finalRegex = `(${matchPackage}|${matchAlias}|${matchSingle}) lang:go `

        eventLogger.log('GoImportsSearchQueryTransformed')

        return query.replace(goImportsRegex, finalRegex)
    }
    return query
}
