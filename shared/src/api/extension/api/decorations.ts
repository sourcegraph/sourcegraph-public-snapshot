import { uniqueId } from 'lodash'

export const createDecorationType = () => ({ key: uniqueId('TextDocumentDecorationType') })
