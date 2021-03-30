declare module '@storybook/addon-console' {
    export declare const withConsole: () => (storyFn: any) => (context: StoryContext) => React.ReactElement
}

declare interface Window {
    MonacoEnvironment: {
        getWorkerUrl(moduleId: string, label: string): string
    }
}
