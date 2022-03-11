import { mdiPoll } from '@mdi/js'
import React from 'react'

import { AccessibleIcon, Icon } from '@sourcegraph/wildcard'

export const CodeInsightsIcon: AccessibleIcon = props => <Icon svgPath={mdiPoll} {...props} />
