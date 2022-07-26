import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import * as H from 'history'

export const [locationField, updateLocation] = createUpdateableField<H.Location>()
