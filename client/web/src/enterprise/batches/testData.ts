import { subSeconds } from 'date-fns'

import {
    BatchSpecListFields,
    BatchSpecSource,
    BatchSpecState,
    BatchSpecWorkspaceFileResult,
} from '../../graphql-operations'

const COMMON_NODE_FIELDS = {
    __typename: 'BatchSpec',
    createdAt: subSeconds(new Date(), 30).toISOString(),
    startedAt: subSeconds(new Date(), 25).toISOString(),
    finishedAt: new Date().toISOString(),
    originalInput: 'name: super-cool-spec',
    description: {
        __typename: 'BatchChangeDescription',
        name: 'super-cool-spec',
    },
    source: BatchSpecSource.LOCAL,
    namespace: {
        url: '/users/courier-new',
        namespaceName: 'courier-new',
    },
    creator: {
        username: 'courier-new',
    },
    files: null,
} as const

export const successNode = (id: string): BatchSpecListFields => ({
    ...COMMON_NODE_FIELDS,
    id,
    state: BatchSpecState.COMPLETED,
})

export const NODES: BatchSpecListFields[] = [
    { ...COMMON_NODE_FIELDS, id: 'id1', state: BatchSpecState.QUEUED },
    { ...COMMON_NODE_FIELDS, id: 'id2', state: BatchSpecState.PROCESSING },
    successNode('id3'),
    { ...COMMON_NODE_FIELDS, id: 'id4', state: BatchSpecState.FAILED },
    { ...COMMON_NODE_FIELDS, id: 'id5', state: BatchSpecState.CANCELING },
    { ...COMMON_NODE_FIELDS, id: 'id6', state: BatchSpecState.CANCELED },
    {
        ...COMMON_NODE_FIELDS,
        state: BatchSpecState.COMPLETED,
        source: BatchSpecSource.REMOTE,
        id: 'id7',
        originalInput: `name: super-cool-spec
description: doing something super interesting

on:
    - repository: github.com/foo/bar
`,
        description: {
            __typename: 'BatchChangeDescription',
            name: 'remote-super-cool-spec',
        },
        files: {
            totalCount: 2,
            pageInfo: {
                endCursor: null,
                hasNextPage: false,
            },
            nodes: [
                {
                    id: 'fileId1',
                    name: 'test.sh',
                    binary: false,
                    byteSize: 12,
                },
                {
                    id: 'fileId2',
                    name: 'src-cli',
                    binary: true,
                    byteSize: 19000,
                },
            ],
        },
    },
    {
        ...COMMON_NODE_FIELDS,
        state: BatchSpecState.COMPLETED,
        source: BatchSpecSource.LOCAL,
        id: 'id8',
        originalInput: `name: super-cool-spec
description: doing something super interesting

on:
    - repository: github.com/foo/bar
`,
        description: {
            __typename: 'BatchChangeDescription',
            name: 'local-super-cool-spec',
        },
        files: {
            totalCount: 1,
            pageInfo: {
                endCursor: null,
                hasNextPage: false,
            },
            nodes: [
                {
                    id: 'fileId3',
                    name: 'test.sh',
                    binary: false,
                    byteSize: 12,
                },
            ],
        },
    },
]

export const MOCK_HIGHLIGHTED_FILES: BatchSpecWorkspaceFileResult = {
    __typename: 'Query',
    node: {
        __typename: 'BatchSpecWorkspaceFile',
        id: 'fileId1',
        name: 'test.sh',
        binary: false,
        byteSize: 12,
        highlight: {
            __typename: 'HighlightedFile',
            aborted: false,
            html: `<table>
            <tbody>
              <tr>
                <th class="line" data-line="1"></th>
                <td class="code">
                  <div>
                    <span class="hl-source hl-shell hl-bash">
                      <span class="hl-comment hl-line hl-number-sign hl-shell">
                        <span
                          class="hl-punctuation hl-definition hl-comment hl-begin hl-shell"
                        >
                          #
                        </span>
                      </span>
                      <span class="hl-comment hl-line hl-number-sign hl-shell">
                        !/usr/bin/env bash
                      </span>
                      <span class="hl-comment hl-line hl-number-sign hl-shell"> </span>
                    </span>
                  </div>
                </td>
              </tr>

              <tr>
                <th class="line" data-line="2"></th>
                <td class="code">
                  <div>
                    <span class="hl-source hl-shell hl-bash"> </span>
                  </div>
                </td>
              </tr>

              <tr>
                <th class="line" data-line="3"></th>
                <td class="code">
                   <div>
                      <span class="hl-source hl-shell hl-bash">
                         <span class="hl-meta hl-function-call hl-shell">
                            <span class="hl-support hl-function hl-echo hl-shell">
                               echo
                            </span>
                         </span>

                         <span class="hl-meta hl-function-call hl-arguments hl-shell">
                            Hello World
                         </span>
                         <span class="hl-keyword hl-operator hl-logical hl-pipe hl-shell">|</span>

                         <span class="hl-meta hl-function-call hl-shell">
                            <span class="hl-variable hl-function hl-shell">tee</span>
                         </span>
                         <span class="hl-meta hl-function-call hl-arguments hl-shell">
                            <span class="hl-variable hl-parameter hl-option hl-shell">
                               <span class="hl-punctuation hl-definition hl-parameter hl-shell">-</span>
                               a
                            </span>
                            <span class="hl-string hl-quoted hl-double hl-shell">
                               <span class="hl-punctuation hl-definition hl-string hl-begin hl-shell">"</span>
                               <span class="hl-meta hl-group hl-expansion hl-command hl-parens hl-shell">
                                  <span class="hl-punctuation hl-definition hl-variable hl-shell">$</span>
                                  <span class="hl-punctuation hl-section hl-parens hl-begin hl-shell">(</span>
                                  <span class="hl-meta hl-function-call hl-shell">
                                     <span class="hl-variable hl-function hl-shell">find</span>
                                  </span>
                                  <span class="hl-meta hl-function-call hl-arguments hl-shell">
                                     <span class="hl-variable hl-parameter hl-option hl-shell">
                                        <span class="hl-punctuation hl-definition hl-parameter hl-shell">-</span>
                                        name
                                     </span>
                                     README.md
                                  </span>
                                  <span class="hl-punctuation hl-section hl-parens hl-end hl-shell">)</span>
                               </span>

                            <span class="hl-punctuation hl-definition hl-string hl-end hl-shell">"</span>
                         </span>
                      </span>
                   </span>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>`,
        },
    },
}
