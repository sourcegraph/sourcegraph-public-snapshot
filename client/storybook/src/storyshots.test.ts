import initStoryshots from '@storybook/addon-storyshots'

import { ENVIRONMENT_CONFIG } from './environment-config'

ENVIRONMENT_CONFIG.STORIES_GLOB = 'client/web/src/nav/GlobalNavbar.story.tsx'

initStoryshots({
    configPath: __dirname,
    framework: 'react',
    suite: 'Storybook (snapshots)',
    // storyNameRegex: /^web\/nav\/GlobalNav/,
})
