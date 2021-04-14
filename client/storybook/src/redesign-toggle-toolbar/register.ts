import addons, { types } from '@storybook/addons'

import { RedesignToggleStorybook } from './RedesignToggleStorybook'

/**
 * Custom toolbar which renders button to toggle redesign theme global CSS class.
 */
addons.register('sourcegraph/redesign-toggle-toolbar', () => {
    addons.add('sourcegraph/redesign-toggle-toolbar', {
        title: 'Redesign toggle toolbar',
        type: types.TOOL,
        match: ({ viewMode }) => viewMode === 'story' || viewMode === 'docs',
        render: RedesignToggleStorybook,
    })
})
