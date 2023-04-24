import { MockedProvider } from '@apollo/client/testing'
import { cleanup } from '@testing-library/react'
import { EMPTY, NEVER } from 'rxjs'
import sinon from 'sinon'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { RepositoryFields } from '../../graphql-operations'

import { Props, TreePage } from './TreePage'

describe('TreePage', () => {
    afterEach(cleanup)

    const repoDefaults = (): RepositoryFields => ({
        id: 'repo-id',
        name: 'repo name',
        url: 'http://repo.url.example.com',
        description: 'Awesome for testing',
        viewerCanAdminister: false,
        isFork: false,
        externalURLs: [],
        externalRepository: {
            serviceType: 'REPO SERVICE TYPE',
            serviceID: 'repo-service-id',
        },
        defaultBranch: {
            displayName: 'Default Branch Display Name',
            abbrevName: 'def-branch-abbr',
        },
        metadata: [],
    })

    const treePagePropsDefaults = (repositoryFields: RepositoryFields): Props => ({
        repo: repositoryFields,
        repoName: 'test repo',
        filePath: '',
        commitID: 'asdf1234',
        revision: 'asdf1234',
        isSourcegraphDotCom: false,
        settingsCascade: {
            subjects: null,
            final: null,
        },
        extensionsController: null,
        platformContext: {
            settings: NEVER,
            updateSettings: () => Promise.reject(new Error('updateSettings not implemented')),
            getGraphQLClient: () => Promise.reject(new Error('getGraphQLClient not implemented')),
            requestGraphQL: () => EMPTY,
            createExtensionHost: () => Promise.reject(new Error('createExtensionHost not implemented')),
            urlToFile: () => '',
            sourcegraphURL: 'https://sourcegraph.com',
            clientApplication: 'sourcegraph',
        },
        telemetryService: NOOP_TELEMETRY_SERVICE,
        codeIntelligenceEnabled: false,
        batchChangesExecutionEnabled: false,
        batchChangesEnabled: false,
        batchChangesWebhookLogsEnabled: false,
        selectedSearchContextSpec: '',
        setBreadcrumb: sinon.spy(),
        useBreadcrumb: sinon.spy(),
        ownEnabled: false,
    })

    describe('repo page', () => {
        it('displays a page that is not a fork', () => {
            const repo = repoDefaults()
            repo.isFork = false
            const props = treePagePropsDefaults(repo)
            const result = renderWithBrandedContext(
                <MockedProvider>
                    <TreePage {...props} />
                </MockedProvider>
            )
            expect(result.queryByTestId('repo-fork-badge')).not.toBeInTheDocument()
            // check for validity that repo header renders
            expect(result.queryByTestId('repo-header')).toBeInTheDocument()
        })

        it('displays a page that is a fork', () => {
            const repo = repoDefaults()
            repo.isFork = true
            const props = treePagePropsDefaults(repo)
            const result = renderWithBrandedContext(
                <MockedProvider>
                    <TreePage {...props} />
                </MockedProvider>
            )
            expect(result.queryByTestId('repo-fork-badge')).toHaveTextContent('Fork')
        })
    })
})
