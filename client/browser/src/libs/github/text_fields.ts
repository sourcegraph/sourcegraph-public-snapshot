import { TextField } from '../code_intelligence/text_fields'
import { ViewResolver } from '../code_intelligence/views'

export const commentTextFieldResolver: ViewResolver<TextField> = {
    selector: '.comment-form-textarea',
    resolveView: element => {
        if (!(element instanceof HTMLTextAreaElement)) {
            return null
        }
        return { element }
    },
}
