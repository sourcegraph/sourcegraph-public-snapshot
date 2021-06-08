import { storiesOf } from '@storybook/react'
import React from 'react'
import { MemoryRouter } from 'react-router'

import { CaseInsensitiveFuzzySearch } from '../../fuzzyFinder/CaseInsensitiveFuzzySearch'
import { SearchValue } from '../../fuzzyFinder/FuzzySearch'
import { WebStory } from '../WebStory'

import { Ready } from './FuzzyFinder'
import { FuzzyModal } from './FuzzyModal'

let query = 'client'
const filenames = [
    '.buildkite/hooks/post-checkout',
    '.buildkite/hooks/pre-command',
    '.buildkite/pipeline.async.yml',
    '.buildkite/pipeline.codeintel.yml',
    '.buildkite/pipeline.e2e.yml',
    '.buildkite/pipeline.qa.yml',
    '.buildkite/updater/is-tip-of-main.sh',
    '.buildkite/updater/pipeline.update-trigger.yaml',
    '.buildkite/updater/trigger-if-tip-of-main.sh',
    '.buildkite/vagrant-run.sh',
    '.dockerignore',
    '.editorconfig',
    '.eslintignore',
    '.eslintrc.js',
    '.gitattributes',
    '.github/CODEOWNERS',
    '.github/ISSUE_TEMPLATE/bug_report.md',
    '.github/ISSUE_TEMPLATE/customer_feedback.md',
    '.github/ISSUE_TEMPLATE/design_debt.md',
    '.github/ISSUE_TEMPLATE/docs-issue.md',
    'client/.github/ISSUE_TEMPLATE/feature_request.md.github/ISSUE_TEMPLATE/feature_request.md.github/ISSUE_TEMPLATE/feature_request.md.github/ISSUE_TEMPLATE/feature_request.md.github/ISSUE_TEMPLATE/feature_request.md',
    '.github/ISSUE_TEMPLATE/flaky_test.md',
    '.github/ISSUE_TEMPLATE/question.md',
    '.github/ISSUE_TEMPLATE/request_patch_release.md',
    '.github/ISSUE_TEMPLATE/tracking_issue.md',
    '.github/ISSUE_TEMPLATE/wildcard_proposal.md',
    '.github/PULL_REQUEST_TEMPLATE.md',
    '.github/PULL_REQUEST_TEMPLATE/browser_extension.md',
    '.github/PULL_REQUEST_TEMPLATE/pull_request_template.md',
    '.github/teams.yml',
    '.github/workflows/CODENOTIFY',
    '.github/workflows/automerge.yml',
    '.github/workflows/batches-notify.yml',
    '.github/workflows/codenotify.yml',
    '.github/workflows/codeql.yml',
    '.github/workflows/container-scanning.yml',
    '.github/workflows/label-notify.yml',
    '.github/workflows/licenses-check.yml',
    '.github/workflows/licenses-update.yml',
    '.github/workflows/lsif.yml',
    '.github/workflows/progress.yml',
    '.github/workflows/renovate-downstream.json',
    '.github/workflows/renovate-downstream.yml',
    '.github/workflows/resources-report.yml',
    '.github/workflows/reviewdog.yml',
    '.github/workflows/team-labeler.yml',
    '.github/workflows/tracking-issue.yml',
    '.gitignore',
    '.gitmodules',
    '.golangci.yml',
    '.graphqlrc.yml',
    '.mailmap',
    '.mocharc.js',
    '.nvmrc',
    '.percy.yml',
    '.prettierignore',
    '.stylelintignore',
    '.stylelintrc.json',
    '.tool-versions',
    '.vscode/extensions.json',
    '.vscode/launch.json',
    '.vscode/settings.json',
    '.vscode/tasks.json',
    '.yarnrc',
    'CHANGELOG.md',
    'CODENOTIFY',
    'CONTRIBUTING.md',
    'LICENSE',
    'LICENSE.apache',
    'LICENSE.enterprise',
    'README.md',
    'SECURITY.md',
    'babel.config.js',
    'client/README.md',
    'client/branded/.eslintignore',
    'client/branded/.eslintrc.js',
    'client/branded/.stylelintrc.json',
    'client/branded/README.md',
    'client/branded/babel.config.js',
    'client/branded/jest.config.js',
    'client/branded/package.json',
    'client/branded/src/components/BrandedStory.tsx',
    'client/branded/src/components/CodeSnippet.tsx',
    'client/branded/src/components/Form.tsx',
    'client/branded/src/components/LoaderInput.scss',
    'client/branded/src/components/LoaderInput.story.tsx',
    'client/branded/src/components/LoaderInput.test.tsx',
    'client/branded/src/components/LoaderInput.tsx',
    'client/branded/src/components/SourcegraphLogo.tsx',
    'client/branded/src/components/Toggle.scss',
    'client/branded/src/components/Toggle.story.tsx',
    'client/branded/src/components/Toggle.test.tsx',
    'client/branded/src/components/Toggle.tsx',
    'client/branded/src/components/ToggleBig.scss',
    'client/branded/src/components/ToggleBig.story.tsx',
    'client/branded/src/components/ToggleBig.test.tsx',
    'client/branded/src/components/ToggleBig.tsx',
    'client/branded/src/components/__snapshots__/LoaderInput.test.tsx.snap',
    'client/branded/src/components/__snapshots__/Toggle.test.tsx.snap',
    'client/branded/src/components/__snapshots__/ToggleBig.test.tsx.snap',
]
const searchValues: SearchValue[] = filenames.map(filename => ({ text: filename }))
const fuzzy = new CaseInsensitiveFuzzySearch(searchValues)
const fsm: Ready = { key: 'ready', fuzzy }
const defaultProps = {
    commitID: 'commitID',
    repoName: 'repoName',
    downloadFilenames: () => Promise.resolve(filenames),
    fsm,
    setFsm: () => {},
    focusIndex: 0,
    setFocusIndex: () => {},
    maxResults: 100,
    increaseMaxResults: () => {},
    isVisible: true,
    onClose: () => {},
    query,
    setQuery: (newQuery: string): void => {
        query = newQuery
    },
    caseInsensitiveFileCountThreshold: 100,
}
const { add } = storiesOf('web/FuzzyFinder', module).addDecorator(story => (
    <MemoryRouter initialEntries={[{ pathname: '/' }]}>
        <WebStory>{() => story()}</WebStory>
    </MemoryRouter>
))

add('Ready', () => <FuzzyModal {...defaultProps} />)
