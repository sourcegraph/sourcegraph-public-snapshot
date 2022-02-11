import { Validator } from '../../../../../../../form/hooks/useField'

export const validRegexp: Validator<string> = (value = '') => {
    if (value.trim() === '') {
        return
    }

    try {
        new RegExp(value)

        return
    } catch {
        return 'Must be a valid regular expression string'
    }
}
