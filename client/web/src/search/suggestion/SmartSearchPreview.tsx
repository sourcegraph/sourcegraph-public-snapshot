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
import { tap } from 'rxjs/operators'

import { SearchPatternType } from '../../../../shared/src/graphql-operations'
import { LATEST_VERSION, aggregateStreamingSearch } from '../../../../shared/src/search/stream'
import { SearchMode } from '@sourcegraph/shared/src/search'

import { useNavbarQueryState } from '../../stores'

import styles from './QuerySuggestion.module.scss'

interface SmartSearchPreviewProps {
    // alert: Required<AggregateStreamingSearchResults>['alert'] | undefined
    // onDisableSmartSearch: () => void
}

export const SmartSearchPreview: React.FunctionComponent<React.PropsWithChildren<SmartSearchPreviewProps>> = () => {
    const options = {
        version: LATEST_VERSION,
        patternType: SearchPatternType.standard,
        caseSensitive: false,
        trace: undefined,
        searchMode: SearchMode.SmartSearch,
    }

    //TODO: How to grab line 50
    //TODO: How to change SmartSearch setting from here
    const searchQuery = useNavbarQueryState(state => state.searchQueryFromURL)
    const smartSearchResults = aggregateStreamingSearch(of(searchQuery), options)
        .pipe(tap(obj => console.log('Results: ', obj)))
        .subscribe()
    console.log(smartSearchResults)

    //line 50 is results.alert I need, how to capture this value in right order?
    return <div className={styles.root}></div>
}
