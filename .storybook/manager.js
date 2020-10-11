import { addons } from '@storybook/addons'
import * as themes from './themes'

addons.setConfig({
  // We have to use dark by default because
  // when the addon is not running (like in Chromatic)
  // our CSS uses dark by default.
  theme: themes.dark,
})
