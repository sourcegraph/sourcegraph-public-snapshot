
# Wildcard combobox primitives 

These components are simply wrappers over the [reach-ui combobox primitives](https://reach.tech/combobox#comboboxpopover).
We do use these reach-ui primitives in order to have covered all a11 peculiarities in how combobox UI should work. 

For more details we suggest take a look at `Combobox.story.tsx` file (storybook demos) but in a nutshell combobox
package exposes low-level UI blocks that 
- Enforce Sourcegraph-like styles 
- Enforce all important accessibility aspects around keyboard and screen-reader navigations.

Simple example

```tsx
import { Combobox, ComboboxInput, ComboboxPopover, ComboboxList, ComboboxOption } from '@sourcegraph/wildcard' 

function RepositoriesCombobox() {
    return (
      <Combobox aria-label="Choose a repo" style={{ maxWidth: '20rem' }}>
        <ComboboxInput
          label="Repository"
          placeholder="Start type..."
          message="You need to specify repo name (github.com/sg/sg) and then pick one of the suggestions items."
        />

        <ComboboxPopover>
          <ComboboxList>
            <ComboboxOption value="github.com/sourcegraph/sourcegraph" />
            <ComboboxOption value="github.com/sourcegraph/about" />
            <ComboboxOption value="github.com/sourcegraph/deploy" />
            <ComboboxOption value="github.com/sourcegraph/handbook" />
          </ComboboxList>
        </ComboboxPopover>
      </Combobox>
    )
}
```

Where 
- `<Combobox />` is a main root component for all other combobox compound components. It stores all important 
information about combobox state (without this component you can't use combobox UI)
- `<ComboboxInput />` is an input component, you can override component implementation with `as` prop, but by 
default we use a standard text `<Input />` component from the wildcard package.
- `<ComboboxPopover />` (*option if you don't want to render suggestions in popover). It adds a popover component where
all suggestions will be rendered, by default it uses standard `Popover` component from the wildcard.
- `<ComboboxList />` is a compound wrapper for the all suggestions.
- `<ComboboxOption />` is a suggestions text component (it implements text highlighting logic by default)

See combobox storybook stories for more examples.
