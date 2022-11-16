import { expect } from '@jest/globals'

import { toBeAriaDisabled } from './toBeAriaDisabled'
import { toBeAriaEnabled } from './toBeAriaEnabled'

expect.extend({
    toBeAriaEnabled,
    toBeAriaDisabled,
})
