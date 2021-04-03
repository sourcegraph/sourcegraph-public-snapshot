import React, { useState } from 'react';
import { createBrowserHistory } from 'history'
import { InsightPage } from '../src/InsightsPage';
import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/out/src/settings/settings';
import { SearchPatternType } from '@sourcegraph/shared/out/src/graphql-operations';
import { render } from 'react-dom';

import '../../SourcegraphWebApp.scss';
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/out/src/telemetry/telemetryService';
import { ViewInsightProviderSourceType } from '../src/core/backend';

// const platformContext: PlatformContext = createPlatformContext()
// const extensionsController: ExtensionsController = createExtensionsController(platformContext)
const history = createBrowserHistory();

const MOCK_VIEWS = [
    {
        'id': 'searchInsights.searchInsights.insight.graphQLTypesMigration.insightsPage', 'view': {
            'title': 'Migration to new GraphQL TS types', 'content': [{
                'chart': 'line' as const,
                'data': [{
                    'date': 1595624400000,
                    'Imports of old GQL.* types': 259,
                    'Imports of new graphql-operations types': 7
                }, {
                    'date': 1599253200000,
                    'Imports of old GQL.* types': 190,
                    'Imports of new graphql-operations types': 191
                }, {
                    'date': 1602882000000,
                    'Imports of old GQL.* types': 182,
                    'Imports of new graphql-operations types': 210
                }, {
                    'date': 1606510800000,
                    'Imports of old GQL.* types': 179,
                    'Imports of new graphql-operations types': 256
                }, {
                    'date': 1610139600000,
                    'Imports of old GQL.* types': 139,
                    'Imports of new graphql-operations types': 335
                }, {
                    'date': 1613768400000,
                    'Imports of old GQL.* types': 139,
                    'Imports of new graphql-operations types': 352
                }, {
                    'date': 1617397200000,
                    'Imports of old GQL.* types': 139,
                    'Imports of new graphql-operations types': 362
                }],
                'series': [{
                    'dataKey': 'Imports of old GQL.* types',
                    'name': 'Imports of old GQL.* types',
                    'stroke': 'var(--oc-red-7)',
                    'linkURLs': ['https://sourcegraph.com/search?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24+type%3Adiff+after%3A2020-06-13T00%3A00%3A00%2B03%3A00+before%3A2020-07-25T00%3A00%3A00%2B03%3A00+patternType%3Aregex+case%3Ayes+%5C*%5Csas%5CsGQL', 'https://sourcegraph.com/search?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24+type%3Adiff+after%3A2020-07-25T00%3A00%3A00%2B03%3A00+before%3A2020-09-05T00%3A00%3A00%2B03%3A00+patternType%3Aregex+case%3Ayes+%5C*%5Csas%5CsGQL', 'https://sourcegraph.com/search?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24+type%3Adiff+after%3A2020-09-05T00%3A00%3A00%2B03%3A00+before%3A2020-10-17T00%3A00%3A00%2B03%3A00+patternType%3Aregex+case%3Ayes+%5C*%5Csas%5CsGQL', 'https://sourcegraph.com/search?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24+type%3Adiff+after%3A2020-10-17T00%3A00%3A00%2B03%3A00+before%3A2020-11-28T00%3A00%3A00%2B03%3A00+patternType%3Aregex+case%3Ayes+%5C*%5Csas%5CsGQL', 'https://sourcegraph.com/search?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24+type%3Adiff+after%3A2020-11-28T00%3A00%3A00%2B03%3A00+before%3A2021-01-09T00%3A00%3A00%2B03%3A00+patternType%3Aregex+case%3Ayes+%5C*%5Csas%5CsGQL', 'https://sourcegraph.com/search?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24+type%3Adiff+after%3A2021-01-09T00%3A00%3A00%2B03%3A00+before%3A2021-02-20T00%3A00%3A00%2B03%3A00+patternType%3Aregex+case%3Ayes+%5C*%5Csas%5CsGQL', 'https://sourcegraph.com/search?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24+type%3Adiff+after%3A2021-02-20T00%3A00%3A00%2B03%3A00+before%3A2021-04-03T00%3A00%3A00%2B03%3A00+patternType%3Aregex+case%3Ayes+%5C*%5Csas%5CsGQL']
                }, {
                    'dataKey': 'Imports of new graphql-operations types',
                    'name': 'Imports of new graphql-operations types',
                    'stroke': 'var(--oc-blue-7)',
                    'linkURLs': ['https://sourcegraph.com/search?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24+type%3Adiff+after%3A2020-06-13T00%3A00%3A00%2B03%3A00+before%3A2020-07-25T00%3A00%3A00%2B03%3A00+patternType%3Aregexp+case%3Ayes+%2Fgraphql-operations%27', 'https://sourcegraph.com/search?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24+type%3Adiff+after%3A2020-07-25T00%3A00%3A00%2B03%3A00+before%3A2020-09-05T00%3A00%3A00%2B03%3A00+patternType%3Aregexp+case%3Ayes+%2Fgraphql-operations%27', 'https://sourcegraph.com/search?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24+type%3Adiff+after%3A2020-09-05T00%3A00%3A00%2B03%3A00+before%3A2020-10-17T00%3A00%3A00%2B03%3A00+patternType%3Aregexp+case%3Ayes+%2Fgraphql-operations%27', 'https://sourcegraph.com/search?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24+type%3Adiff+after%3A2020-10-17T00%3A00%3A00%2B03%3A00+before%3A2020-11-28T00%3A00%3A00%2B03%3A00+patternType%3Aregexp+case%3Ayes+%2Fgraphql-operations%27', 'https://sourcegraph.com/search?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24+type%3Adiff+after%3A2020-11-28T00%3A00%3A00%2B03%3A00+before%3A2021-01-09T00%3A00%3A00%2B03%3A00+patternType%3Aregexp+case%3Ayes+%2Fgraphql-operations%27', 'https://sourcegraph.com/search?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24+type%3Adiff+after%3A2021-01-09T00%3A00%3A00%2B03%3A00+before%3A2021-02-20T00%3A00%3A00%2B03%3A00+patternType%3Aregexp+case%3Ayes+%2Fgraphql-operations%27', 'https://sourcegraph.com/search?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24+type%3Adiff+after%3A2021-02-20T00%3A00%3A00%2B03%3A00+before%3A2021-04-03T00%3A00%3A00%2B03%3A00+patternType%3Aregexp+case%3Ayes+%2Fgraphql-operations%27']
                }],
                'xAxis': {'dataKey': 'date', 'type': 'number', 'scale': 'time'}
            }]
        }, 'source': ViewInsightProviderSourceType.Extension
    },
    {
        'id': 'codeStatsInsights.languages.insightsPage', 'view': {
            'title': 'Language usage', 'content': [{
                'chart': 'pie',
                'pies': [{
                    'data': [{
                        'name': 'Go',
                        'totalLines': 363432,
                        'fill': '#00ADD8',
                        'linkURL': 'https://sourcegraph.com/stats?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24'
                    }, {
                        'name': 'HTML',
                        'totalLines': 224961,
                        'fill': '#e34c26',
                        'linkURL': 'https://sourcegraph.com/stats?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24'
                    }, {
                        'name': 'TypeScript',
                        'totalLines': 155381,
                        'fill': '#2b7489',
                        'linkURL': 'https://sourcegraph.com/stats?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24'
                    }, {
                        'name': 'Markdown',
                        'totalLines': 46675,
                        'fill': '#083fa1',
                        'linkURL': 'https://sourcegraph.com/stats?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24'
                    }, {
                        'name': 'YAML',
                        'totalLines': 25412,
                        'fill': '#cb171e',
                        'linkURL': 'https://sourcegraph.com/stats?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24'
                    }, {
                        'name': 'Other',
                        'totalLines': 56846,
                        'fill': 'gray',
                        'linkURL': 'https://sourcegraph.com/stats?q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24'
                    }], 'dataKey': 'totalLines', 'nameKey': 'name', 'fillKey': 'fill', 'linkURLKey': 'linkURL'
                }]
            }]
        }, 'source': 'Extension'
    }
]

export function App() {
    const [patternType, setPatterType] = useState(SearchPatternType.literal);
    const [caseSensitive, setCaseSensitive] = useState(false);

    return (
        <InsightPage
            versionContext={undefined}
            telemetryService={NOOP_TELEMETRY_SERVICE}
            copyQueryButton={false}
            caseSensitive={caseSensitive}
            setCaseSensitivity={setCaseSensitive}
            setPatternType={setPatterType}
            patternType={patternType}
            settingsCascade={EMPTY_SETTINGS_CASCADE}
            globbing={false}
            location={history.location}
            views={MOCK_VIEWS}
            history={history}
            /* eslint-disable-next-line @typescript-eslint/ban-ts-comment */
            // @ts-ignore
            extensionsController={null}/>
    );
}

render(<App/>, document.querySelector('#root'));

