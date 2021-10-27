import { storiesOf } from '@storybook/react'
import React from 'react'

import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'

import { WebStory } from '../../../components/WebStory'

import { CreateBatchChangePage } from './CreateBatchChangePage'

const { add } = storiesOf('web/batches/CreateBatchChangePage', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('experimental execution disabled', () => (
    <WebStory>
        {props => <CreateBatchChangePage headingElement="h1" {...props} settingsCascade={EMPTY_SETTINGS_CASCADE} />}
    </WebStory>
))

// TODO: Disabled because it has a infinite render loop and caused flakes in CI.
// const FIXTURE_ORG: SettingsOrgSubject = {
//     __typename: 'Org',
//     name: 'sourcegraph',
//     displayName: 'Sourcegraph',
//     id: 'a',
//     viewerCanAdminister: true,
// }

// const FIXTURE_USER: SettingsUserSubject = {
//     __typename: 'User',
//     username: 'alice',
//     displayName: 'alice',
//     id: 'b',
//     viewerCanAdminister: true,
// }
// const batchSpec: BatchSpecWithWorkspacesFields = {
//     __typename: 'BatchSpec',
//     allowIgnored: false,
//     allowUnsupported: false,
//     id: 'SUPERID',
//     importingChangesets: { totalCount: 0, nodes: [] },
//     originalInput: 'name: ',
//     workspaceResolution: {
//         failureMessage: null,
//         state: BatchSpecWorkspaceResolutionState.FAILED,
//         workspaces: { nodes: [] },
//     },
// }

// const createBatchSpecFromRaw = () => of(batchSpec)

// add('experimental execution enabled', () => (
//     <WebStory>
//         {props => (
//             <CreateBatchChangePage
//                 createBatchSpecFromRaw={createBatchSpecFromRaw}
//                 headingElement="h1"
//                 {...props}
//                 settingsCascade={{
//                     ...EMPTY_SETTINGS_CASCADE,
//                     final: {
//                         experimentalFeatures: { batchChangesExecution: true },
//                     },
//                     subjects: [
//                         { subject: FIXTURE_ORG, settings: { a: 1 }, lastID: 1 },
//                         { subject: FIXTURE_USER, settings: { b: 2 }, lastID: 2 },
//                     ],
//                 }}
//             />
//         )}
//     </WebStory>
// ))
