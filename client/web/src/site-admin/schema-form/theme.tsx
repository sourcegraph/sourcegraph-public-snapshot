import { mdiPlus } from '@mdi/js'
import { RegistryWidgetsType, WidgetProps, FieldTemplateProps } from '@rjsf/utils'

import { Button, Checkbox, Icon, Input } from '@sourcegraph/wildcard'

const AddButton = () => {
    return (
        <Button variant="success">
            <Icon svgPath={mdiPlus} aria-hidden={true} /> Add
        </Button>
    )
}

const InputWidget: React.FC<WidgetProps> = (props: WidgetProps) => {
    const { ...rest } = props

    return <Input {...rest} />
}

const customWidgets: RegistryWidgetsType = {
    CheckboxWidget: Checkbox,
    ButtonWidget: Button,
    TextWidget: InputWidget,
}

export const theme = {
    widgets: customWidgets,
    templates: {
        ButtonTemplates: { AddButton },
    },
}
