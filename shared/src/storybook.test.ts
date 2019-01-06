import initStoryshots from '@storybook/addon-storyshots'
import * as path from 'path'

initStoryshots({ framework: 'react', configPath: path.resolve(__dirname, '..', '.storybook') })
