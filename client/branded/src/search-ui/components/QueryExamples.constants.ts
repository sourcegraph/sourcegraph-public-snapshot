/** Static query examples */

import { isDefined } from '@sourcegraph/common'

import type { QueryExamplesSection } from './useQueryExamples'

export const exampleQueryColumns = [
    [
        {
            title: 'Find usage examples',
            queryExamples: [
                { query: 'context.WithCancel lang:go' },
                { query: '<Suspense lang:typescript' },
                { query: 'readFileSync lang:javascript' },
                { query: 'import torch lang:python' },
            ],
        },
    ],
    [
        {
            title: 'Find TODOs in a repository',
            queryExamples: [{ query: 'repo:facebook/react TODO' }],
        },
        {
            title: 'See API usage and changes over time',
            queryExamples: [{ query: 'repo:pytorch/pytorch type:diff is_cpu' }],
        },
    ],
]

export const basicSyntaxColumns = (
    fileName: string,
    singleRepoExample: string,
    orgReposExample: string | undefined,
    keywordSearch: boolean
): QueryExamplesSection[][] =>
    keywordSearch
        ? [
              [
                  {
                      title: 'Search in files, paths, and repository-names',
                      queryExamples: [
                          { query: 'test server', helperText: '(both terms anywhere)', productStatus: 'new' },
                          { query: '"Error 1001"', helperText: '(specific string)', productStatus: 'new' },
                          {
                              query: '"\\"Error 1001\\""',
                              helperText: '(specific string containing quotes)',
                              productStatus: 'new',
                          },
                          { query: 'foo OR bar' },
                          { query: '/open(File|Dir)/', helperText: '(regular expression)' },
                      ],
                  },
                  {
                      title: 'Search in commit diffs',
                      queryExamples: [{ query: 'type:diff after:1week fix' }, { query: 'type:diff author:alice add' }],
                  },
              ],
              [
                  {
                      title: 'Filter by...',
                      queryExamples: [
                          { query: `file:${fileName} foo` },
                          { query: `repo:${singleRepoExample}` },
                          orgReposExample
                              ? { query: `repo:${orgReposExample}`, helperText: '(all repositories in org)' }
                              : null,
                          { query: 'lang:javascript' },
                      ].filter(isDefined),
                  },
                  {
                      title: 'Advanced',
                      queryExamples: [
                          { query: 'repo:has.description(foo)' },
                          { query: 'file:^some_path file:has.owner(alice)' },
                          { query: 'file:^some_path select:file.owners' },
                          { query: 'file:has.commit.after(1week)' },
                      ],
                  },
              ],
          ]
        : [
              [
                  {
                      title: 'Search in files',
                      queryExamples: [
                          { query: 'fetch(' },
                          { query: 'some error message', helperText: '(no quotes needed)' },
                          { query: 'foo AND bar' },
                          { query: '/open(File|Dir)/', helperText: '(regular expression)' },
                      ],
                  },
                  {
                      title: 'Search in commit diffs',
                      queryExamples: [{ query: 'type:diff after:1week fix' }, { query: 'type:diff author:alice add' }],
                  },
              ],
              [
                  {
                      title: 'Filter by...',
                      queryExamples: [
                          { query: `file:${fileName} foo` },
                          { query: `repo:${singleRepoExample}` },
                          orgReposExample
                              ? { query: `repo:${orgReposExample}`, helperText: '(all repositories in org)' }
                              : null,
                          { query: 'lang:javascript' },
                      ].filter(isDefined),
                  },
                  {
                      title: 'Advanced',
                      queryExamples: [
                          { query: 'repo:has.description(foo)' },
                          { query: 'file:^some_path file:has.owner(alice)' },
                          { query: 'file:^some_path select:file.owners' },
                          { query: 'file:has.commit.after(1week)' },
                      ],
                  },
              ],
          ]
