import { highlight } from "highlight.js";

onmessage = (event) => {
	// TODO(john): pass language
	const result = highlight("go", event.data.textContent);
	postMessage({ innerHtml: result.value, index: event.data.index });
}
