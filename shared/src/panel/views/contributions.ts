import { Unsubscribable } from 'rxjs'
import { ExtensionsControllerProps } from '../../extensions/controller'

export function registerPanelToolbarContributions({
    extensionsController,
}: ExtensionsControllerProps<'services'>): Unsubscribable {
    const isLocationsPanelView = ['def', 'references', 'impl', 'typedef']
        .map(id => `(panel.activeView.id == "${id}")`) // must be explicit about operator precedence
        .join(' || ')

    return extensionsController.services.contribution.registerContributions({
        contributions: {
            actions: [
                {
                    id: 'panel.locations.groupByFile',
                    title: 'Group by file',
                    category: 'Locations (panel)',
                    command: 'updateConfiguration',
                    commandArguments: [
                        ['panel.locations.groupByFile'],
                        // tslint:disable-next-line:no-invalid-template-strings
                        '${!config.panel.locations.groupByFile}',
                        null,
                        'json',
                    ],
                    // tslint:disable-next-line:no-invalid-template-strings
                    actionItem: { label: '${config.panel.locations.groupByFile && "Ungroup" || "Group"} by file' },
                },
            ],
            menus: {
                'panel/toolbar': [
                    {
                        action: 'panel.locations.groupByFile',
                        when: `panel.locations.hasResults && (${isLocationsPanelView})`,
                    },
                ],
            },
        },
    })
}
