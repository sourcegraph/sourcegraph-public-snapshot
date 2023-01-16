// import { MouseEvent, useCallback } from 'react'

// import { mdiArrowRight, mdiChevronDown, mdiChevronUp } from '@mdi/js'

// import { formatSearchParameters, pluralize } from '@sourcegraph/common'
// import { SyntaxHighlightedSearchQuery, smartSearchIconSvgPath } from '@sourcegraph/search-ui'
// import { AggregateStreamingSearchResults, AlertKind } from '@sourcegraph/shared/src/search/stream'
// import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
// import {
//     Link,
//     createLinkUrl,
//     Icon,
//     Collapse,
//     CollapseHeader,
//     CollapsePanel,
//     H2,
//     Text,
//     Button,
// } from '@sourcegraph/wildcard'

// import { SearchPatternType } from '../../graphql-operations'
import { noop } from 'lodash'
import { fromEvent, of, Subscriber, Subscription } from 'rxjs'
import { tap } from 'rxjs/operators'

import { SearchPatternType } from '../../../../shared/src/graphql-operations'
import {
    LATEST_VERSION,
    messageHandlers,
    MessageHandlers,
    observeMessages,
    aggregateStreamingSearch,
    SearchEvent,
    switchAggregateSearchResults,
} from '../../../../shared/src/search/stream'
import { SearchMode } from '@sourcegraph/shared/src/search'

import styles from './QuerySuggestion.module.scss'

interface SmartSearchPreviewProps {
    // alert: Required<AggregateStreamingSearchResults>['alert'] | undefined
    // onDisableSmartSearch: () => void
}

// const processDescription = (description: string): string => {
//     const split = description.split(' ⚬ ')

//     split[0] = split[0][0].toUpperCase() + split[0].slice(1)
//     return split.join(', ')
// }

// const alertContent: { [key in AlertKind]: (queryCount: number) => { title: JSX.Element; subtitle: JSX.Element } } = {
//     'smart-search-additional-results': (queryCount: number) => ({
//         title: (
//             <>
//                 <b>Smart Search</b> is also showing <b>additional results</b>.
//             </>
//         ),
//         subtitle: (
//             <>
//                 Smart Search added results for the following similar {pluralize('query', queryCount, 'queries')} that
//                 might interest you:
//             </>
//         ),
//     }),
//     'smart-search-pure-results': (queryCount: number) => ({
//         title: (
//             <>
//                 <b>Smart Search</b> is showing <b>related results</b> as your query found <b>no results</b>.
//             </>
//         ),
//         subtitle: (
//             <>
//                 To get additional results, Smart Search also ran {pluralize('this', queryCount, 'these')}{' '}
//                 {pluralize('query', queryCount, 'queries')}:
//             </>
//         ),
//     }),
// }

export const SmartSearchPreview: React.FunctionComponent<React.PropsWithChildren<SmartSearchPreviewProps>> = () => {
    const options = {
        version: LATEST_VERSION,
        patternType: SearchPatternType.standard,
        caseSensitive: false,
        trace: undefined,
        searchMode: SearchMode.SmartSearch,
    }

    const findSSResults = () => {
        //TODO: How to grab search query dynamically
        //TODO: How to grab pipe obj results
        //TODO: How to change SmartSearch setting from here
        const results = aggregateStreamingSearch(of('sourcegraph javascript'), options)
            .pipe(tap(obj => console.log('HERE: ', obj)))
            .subscribe()
        console.log(results)
    }
    findSSResults()
    return <div className={styles.root}></div>
}
