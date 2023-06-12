import {
    createConnection, InitializeResult, ProposedFeatures,
} from 'vscode-languageserver/node';

export const startServer = (): void => {
    const connection = createConnection(
        ProposedFeatures.all,
    );

    connection.onInitialize(() => {
        connection.console.log('Received an initialization request');
        const ret: InitializeResult = {
            capabilities: {},
        };
        return ret;
    });

    connection.onInitialized(() => {
        connection.console.log('Server initialized');
    });

    connection.listen();
};
