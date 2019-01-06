import initStoryshots from '@storybook/addon-storyshots'
import path from 'path'

initStoryshots({ framework: 'react', configPath: path.resolve(__dirname, '../.storybook/config.js') })
