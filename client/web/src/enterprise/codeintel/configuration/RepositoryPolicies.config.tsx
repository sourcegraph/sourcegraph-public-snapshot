export enum ConfigTypes {
    Global,
    Local,
}

export const CONFIG_TEXT = {
    [ConfigTypes.Global]: {
        title: 'Global policies',
        deleteConfirm: 'Delete global policy',
    },
    [ConfigTypes.Local]: {
        title: 'Repository-specific policies',
        deleteConfirm: 'Delete policy',
    },
}
