import { renderWithBrandedContext } from '@sourcegraph/wildcard'

import { ExtensionRegistrySidenav } from './ExtensionRegistrySidenav'

describe('ExtensionsQueryInputToolbar', () => {
    test('renders', () => {
        expect(
            renderWithBrandedContext(
                <ExtensionRegistrySidenav
                    selectedCategory="Code analysis"
                    onSelectCategory={() => {}}
                    enablementFilter="all"
                    setEnablementFilter={() => {}}
                    showExperimentalExtensions={true}
                    toggleExperimentalExtensions={() => {}}
                />
            ).asFragment()
        ).toMatchSnapshot()
    })
})
