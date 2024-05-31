import { Extension, StateEffect, StateEffectType, StateField } from '@codemirror/state'
import { EditorView } from '@codemirror/view'

/**
 * A helper function for creating an extension that operates on the value which
 * can be updated via an effect.
 * This is useful in React components where the extension depends on the value
 * of a prop  but that prop is unstable, and especially useful for callbacks.
 * Instead of reconfiguring the editor whenever the value changes (which is
 * apparently not cheap), the extension can be updated via the returned update
 * function or effect.
 *
 * Example:
 *
 * const {onChange} = props;
 * const [onChangeField, setOnChange] = useMemo(() => createUpdateableField(...), [])
 * ...
 * useEffect(() => {
 *   if (editor) {
 *     setOnchange(editor, onChange)
 *   }
 * }, [editor, onChange])
 */
export function createUpdateableField<T>(
    defaultValue: T,
    provider?: (field: StateField<T>) => Extension
): [StateField<T>, (editor: EditorView, newValue: T) => void, StateEffectType<T>] {
    const fieldEffect = StateEffect.define<T>()
    const field = StateField.define<T>({
        create() {
            return defaultValue
        },
        update(value, transaction) {
            const effect = transaction.effects.find((effect): effect is StateEffect<typeof defaultValue> =>
                effect.is(fieldEffect)
            )
            return effect ? effect.value : value
        },
        provide: provider,
    })

    return [field, (editor, newValue) => editor.dispatch({ effects: [fieldEffect.of(newValue)] }), fieldEffect]
}
