package architecture

//go:generate sh -c "dot architecture.dot -Tsvg > architecture.svg"

//go:generate sh -c "yarn"
//go:generate sh -c "../../../../node_modules/.bin/mmdc -i definitions.mermaid -o definitions.svg"
//go:generate sh -c "../../../../node_modules/.bin/mmdc -i references.mermaid -o references.svg"
//go:generate sh -c "../../../../node_modules/.bin/mmdc -i resolve-page.mermaid -o resolve-page.svg"
//go:generate sh -c "../../../../node_modules/.bin/mmdc -i hover.mermaid -o hover.svg"
//go:generate sh -c "../../../../node_modules/.bin/mmdc -i upload.mermaid -o upload.svg"
//go:generate sh -c "../../../../node_modules/.bin/mmdc -i extension-definitions.mermaid -o extension-definitions.svg"
//go:generate sh -c "../../../../node_modules/.bin/mmdc -i extension-references.mermaid -o extension-references.svg"
//go:generate sh -c "../../../../node_modules/.bin/mmdc -i extension-hover.mermaid -o extension-hover.svg"
