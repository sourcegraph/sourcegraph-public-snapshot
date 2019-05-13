import { querySelectorAllOrSelf } from '../../shared/util/dom'
import { TextField } from '../code_intelligence/text_fields'
import { ViewResolver } from '../code_intelligence/views'

export const commentTextFieldResolver: ViewResolver<TextField> = container =>
    [...querySelectorAllOrSelf<HTMLTextAreaElement>(container, '.comment-form-textarea')].map(element => ({ element }))
