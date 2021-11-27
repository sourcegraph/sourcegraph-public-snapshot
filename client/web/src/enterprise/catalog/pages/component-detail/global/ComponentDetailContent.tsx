import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

interface Props extends TelemetryProps {}

export const ComponentDetailContent: React.FunctionComponent<Props> = () => <p>Hello from ComponentDetailContent</p>
