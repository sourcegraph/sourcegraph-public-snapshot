// - given current resource text, original text and LLM output (and maybe tracked range) => gross range where edits must trigger a respin, set of highlights for pending edits, set of highlights for conflicts, merged text for diff mode, set of edits for applying a diff (if any of these are expensive, break this up and cache stuff; not everything needs merged text for diff mode, for example)

function f(): void {
    console.log('f')
}
