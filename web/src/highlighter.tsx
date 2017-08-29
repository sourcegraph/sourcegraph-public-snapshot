import { highlight } from 'highlight.js';

onmessage = event => {
    let result;
    const start = Date.now();
    if (event.data.lang === '') {
        // We don't use highlightAuto b/c it's apparently super slow
        result = { value: event.data.textContent };
    } else {
        result = highlight(event.data.lang, event.data.textContent, true);
    }
    console.trace('syntax highlight duration', (Date.now() - start) / 1000);
    const post = postMessage as any;
    post({ innerHTML: result.value });
};
