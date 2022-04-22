import { MockedResponse } from '@apollo/client/testing'
import { render, RenderResult } from '@testing-library/react'
import * as H from 'history'
import { EMPTY, of } from 'rxjs'
import { HoverThresholdProps } from 'src/repo/RepoContainer'

import { getDocumentNode } from '@sourcegraph/http-client'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContext, PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE, TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { extensionsController } from '@sourcegraph/shared/src/testing/searchTestHelpers'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { ReferencesPanelWithMemoryRouter } from './ReferencesPanel'
import { USE_PRECISE_CODE_INTEL_MOCK, RESOLVE_REPO_REVISION_BLOB_MOCK } from './ReferencesPanel.mocks'
import { USE_PRECISE_CODE_INTEL_FOR_POSITION_QUERY, RESOLVE_REPO_REVISION_BLOB_QUERY } from './ReferencesPanelQueries'

const url = '/github.com/sourcegraph/go-diff/-/blob/diff/diff.go?L16:2&subtree=true#tab=references'
const repoName = 'github.com/sourcegraph/go-diff'
const commit = '9d1f353a285b3094bc33bdae277a19aedabe8b71'
const filePath = 'diff/diff.go'

const mocks: readonly MockedResponse[] = [
    {
        request: {
            query: getDocumentNode(USE_PRECISE_CODE_INTEL_FOR_POSITION_QUERY),
            variables: {
                repository: repoName,
                commit,
                path: filePath,
                line: 15,
                character: 1,
                filter: null,
                firstReferences: 100,
                afterReferences: null,
                firstImplementations: 100,
                afterImplementations: null,
            },
        },
        result: USE_PRECISE_CODE_INTEL_MOCK,
    },
    {
        request: {
            query: getDocumentNode(RESOLVE_REPO_REVISION_BLOB_QUERY),
            variables: {
                repoName,
                filePath,
                revision: '',
            },
        },
        result: RESOLVE_REPO_REVISION_BLOB_MOCK,
    },
]

export const NOOP_SETTINGS_CASCADE = {
    subjects: null,
    final: null,
}

export const NOOP_PLATFORM_CONTEXT: Pick<
    PlatformContext,
    'urlToFile' | 'requestGraphQL' | 'settings' | 'forceUpdateTooltip'
> = {
    requestGraphQL: () => EMPTY,
    urlToFile: () => '',
    settings: of(NOOP_SETTINGS_CASCADE),
    forceUpdateTooltip: () => {},
}

const defaultProps: SettingsCascadeProps &
    PlatformContextProps<'urlToFile' | 'requestGraphQL' | 'settings' | 'forceUpdateTooltip'> &
    TelemetryProps &
    HoverThresholdProps &
    ExtensionsControllerProps &
    ThemeProps = {
    extensionsController,
    telemetryService: NOOP_TELEMETRY_SERVICE,
    settingsCascade: {
        subjects: null,
        final: null,
    },
    platformContext: NOOP_PLATFORM_CONTEXT,
    isLightTheme: false,
}

describe('ReferencesPanel', () => {
    beforeAll(() => {
        window.context = { sourcegraphDotComMode: false } as any
    })

    async function renderReferencesPanel() {
        const fakeExternalHistory = H.createMemoryHistory()
        fakeExternalHistory.push(url)

        const result: RenderResult = render(
            <MockedTestProvider mocks={mocks}>
                <ReferencesPanelWithMemoryRouter
                    {...defaultProps}
                    externalHistory={fakeExternalHistory}
                    externalLocation={fakeExternalHistory.location}
                />
            </MockedTestProvider>
        )
        await waitForNextApolloResponse()
        await waitForNextApolloResponse()
        return result
    }

    it('renders references correctly', async () => {
        const result = await renderReferencesPanel()
        const heading = result.getByText('References')
        expect(heading).toBeVisible()

        expect(result.getByText('label = fmt.Sprintf("orig(%s) new(%s)", fdiff.OrigName, fdiff.NewName)')).toBeVisible()
        expect(
            result.getByText('if err := printFileHeader(&buf, "--- ", d.OrigName, d.OrigTime); err != nil {')
        ).toBeVisible()
    })
})
