import {
    createConnection, InitializeResult, ProposedFeatures,
} from 'vscode-languageserver/node';

export const startServer = (): void => {
    const connection = createConnection(
        ProposedFeatures.all,
    );

    connection.onInitialize(() => {
        connection.console.log('Received an initialization request');
        return {} as InitializeResult;
    });

    connection.onInitialized(() => {
        connection.console.log('Server initialized');
    });

    connection.listen();
};
