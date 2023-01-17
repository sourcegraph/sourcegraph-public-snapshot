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

import { of } from 'rxjs'
import { tap, map } from 'rxjs/operators'

import { SearchPatternType } from '../../../../shared/src/graphql-operations'
import {
    LATEST_VERSION,
    aggregateStreamingSearch,
    AggregateStreamingSearchResults,
} from '../../../../shared/src/search/stream'
import { SearchMode } from '@sourcegraph/shared/src/search'

import { useNavbarQueryState } from '../../stores'

import styles from './QuerySuggestion.module.scss'

interface SmartSearchPreviewProps {
    // alert: Required<AggregateStreamingSearchResults>['alert'] | undefined
    // onDisableSmartSearch: () => void
}

export const SmartSearchPreview: React.FunctionComponent<React.PropsWithChildren<SmartSearchPreviewProps>> = () => {
    let test = null

    //TODO: How to change SmartSearch setting from here
    const searchQuery = useNavbarQueryState(state => state.searchQueryFromURL)
    aggregateStreamingSearch(of(searchQuery), {
        version: LATEST_VERSION,
        patternType: SearchPatternType.standard,
        caseSensitive: false,
        trace: undefined,
        searchMode: SearchMode.SmartSearch,
    })
        .pipe(
            // map(data => {
            //     if (!data.alert) {
            //         return
            //     }
            //     return data.alert
            // })
            tap(obj => {
                console.log('Results', obj), (test = obj)
            })
        )
        .subscribe()
    console.log('END', test)
    return <div className={styles.root}></div>
}
