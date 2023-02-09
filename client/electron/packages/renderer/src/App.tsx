import {LazyCodeMirrorQueryInput} from '@sourcegraph/branded/src/search-ui/experimental';
import {Toggles} from '@sourcegraph/branded';
import {QueryState} from '@sourcegraph/shared/src/search';
import {
  EMPTY_SETTINGS_CASCADE,
  SettingsCascadeOrError,
} from '@sourcegraph/shared/src/settings/settings';
import {createSuggestionsSource} from '@sourcegraph/web/src/search/input/suggestions';

import {useCallback, useMemo, useState} from 'react';
import {Router} from 'react-router-dom';
import {CompatRouter, Routes, Route} from 'react-router-dom-v5-compat';
import {createMemoryHistory} from 'history';

import styles from './App.module.scss';
import {Form} from '@sourcegraph/wildcard';
import './globals.scss';
import classNames from 'classnames';
import {PlatformContext} from '@sourcegraph/shared/src/platform/context';
import {requestGraphQLCommon} from '@sourcegraph/http-client';
import {of} from 'rxjs';

interface Search {
  query: QueryState;
  patternType: string; //SearchPatternType;
  caseSensitive: boolean;
  searchMode: number;
}

const history = createMemoryHistory();
export function Wrapper() {
  return (
    <Router history={history}>
      <CompatRouter>
        <Routes>
          <Route
            path="*"
            element={<App />}
          />
        </Routes>
      </CompatRouter>
    </Router>
  );
}

function App() {
  const instanceURL = 'https://sourcegraph.com';

  const [search, setSearch] = useState<Search>({
    query: {
      changeSource: 0,
      query: '',
    },
    patternType: 'literal',
    caseSensitive: false,
    searchMode: 0,
  });

  const requestGraphQL = useCallback<PlatformContext['requestGraphQL']>(
    args =>
      requestGraphQLCommon({
        ...args,
        baseUrl: instanceURL,
        headers: {
          Accept: 'application/json',
          'Content-Type': 'application/json',
        },
      }),
    [instanceURL],
  );

  const settingsCascade: SettingsCascadeOrError = EMPTY_SETTINGS_CASCADE;

  const platformContext = {
    requestGraphQL,
  };

  const suggestionSource = useMemo(
    () =>
      createSuggestionsSource({
        platformContext,
        authenticatedUser: null,
        fetchSearchContexts: () =>
          of({
            __typename: 'SearchContextConnection',
            totalCount: 0,
            nodes: [],
          }),
        getUserSearchContextNamespaces: () => [],
        isSourcegraphDotCom: true,
      }),
    [platformContext],
  );

  const onSubmit = useCallback((...args) => {
    console.log('onSubmit', ...args);
  }, []);

  return (
    <div className="d-flex flex-row flex-shrink-past-contents">
      <Form
        className="flex-grow-1 flex-shrink-past-contents"
        onSubmit={() => console.log('Form#onSubmit')}
      >
        <div
          data-search-page-input-container={true}
          className={styles.inputContainer}
        >
          <div className={classNames('d-flex', 'flex-grow-1', 'w-100', styles.noDrag)}>
            <LazyCodeMirrorQueryInput
              autoFocus={true}
              patternType={search.patternType}
              interpretComments={false}
              queryState={search.query}
              onChange={(query: QueryState) => {
                console.log('onChange', query);
                setSearch(search => ({...search, query}));
              }}
              onSubmit={onSubmit}
              isLightTheme={true}
              placeholder="Search for code or files..."
              suggestionSource={suggestionSource}
            >
              <Toggles
                patternType={search.patternType}
                caseSensitive={search.caseSensitive}
                setPatternType={(...args) => console.log('setPatternType', ...args)}
                setCaseSensitivity={(...args) => console.log('setCaseSensitivity', ...args)}
                searchMode={search.searchMode}
                setSearchMode={(...args) => console.log('setSearchMode', ...args)}
                settingsCascade={settingsCascade}
                navbarSearchQuery={search.query}
              />
            </LazyCodeMirrorQueryInput>
          </div>
        </div>
      </Form>
    </div>
  );
}
