import idCreator from '../../../util/idCreator'

const nextDecorationType = idCreator('TextDocumentDecorationType')
export const createDecorationType = () => ({ key: nextDecorationType() })
