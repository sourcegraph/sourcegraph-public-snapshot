import * as H from 'history'

import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'

export const [locationField, updateLocation] = createUpdateableField<H.Location | null>(null)
