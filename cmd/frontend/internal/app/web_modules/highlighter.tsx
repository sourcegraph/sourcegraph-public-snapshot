import { highlight } from "highlight.js";

onmessage = (event) => {
	// TODO(john): pass language
	const result = highlight("go", event.data.textContent);
	const post = postMessage as any;
	post({ innerHtml: result.value, index: event.data.index });
}
