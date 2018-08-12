import { DocumentSelector } from '../../src/types/document'

export default require('./package.json').cxp as {
    url: string
    accessToken: string
    root: string
    documentSelector: DocumentSelector
    initializationOptions: {
        mode: string
        configurationCascade: { merged: any }
    }
}
