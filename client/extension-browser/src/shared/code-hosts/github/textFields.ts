import { TextField } from '../ui-kit-legacy-shared/textFields'
import { ViewResolver } from '../ui-kit-legacy-shared/views'

export const commentTextFieldResolver: ViewResolver<TextField> = {
    selector: '.comment-form-textarea',
    resolveView: element => {
        if (!(element instanceof HTMLTextAreaElement)) {
            return null
        }
        return { element }
    },
}
