import { mdiPoll } from '@mdi/js'
import React from 'react'

import { Icon } from '@sourcegraph/wildcard'

export const CodeInsightsIcon: typeof Icon = props => <Icon svgPath={mdiPoll} {...props} />
