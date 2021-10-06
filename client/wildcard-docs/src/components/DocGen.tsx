const docGen = require('react-docgen')

const options = {
    savePropValueAsString: true,
}

// Parse a file for docgen info
const outcome = docGen.parse('@sourcegraph/wildcard/src/components/Button/Button.tsx', options)

console.log(outcome)

export const DocGen = () => <div>JSON.stringify(outcome)</div>
