import { of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import type { Panel, TabbedPanelContent } from './TabbedPanelContent'

export const panels: Panel[] = [
    {
        id: 'panel_1',
        title: 'Panel 1',
        element: 'Panel 1',
    },
    {
        id: 'panel_2',
        title: 'Panel 2',
        element: 'Panel 2',
    },
    {
        id: 'panel_3',
        title: 'Panel 3',
        element: 'Panel 3',
    },
]

export const panelProps: React.ComponentPropsWithoutRef<typeof TabbedPanelContent> = {
    repoName: 'git://github.com/foo/bar',
    fetchHighlightedFileLineRanges: () => of([]),
    platformContext: {} as any,
    settingsCascade: { subjects: null, final: null },
    telemetryService: NOOP_TELEMETRY_SERVICE,
}
