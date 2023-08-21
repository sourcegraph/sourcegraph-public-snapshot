# Cody Quickstart

<p class="subtitle">In this quickstart guide, you'll learn how to use Cody once you have installed the extension in your VS Code editor. Here you will:</p>

- Try the `Generate Unit Tests` command
- Ask Cody to pull information from the documentation
- Ask Cody to write context-aware code

## Prerequisites

- Make sure you have the [Cody extension installed](overview/install-vscode.md) in your VS Code editor
- You have enabled an instance for [Cody from your Sourcegraph.com](overview/cody-with-sourcegraph.md) account
- You have a project open in VS Code that Cody has access to via Sourcegraph

## Getting started with Cody extension and commands

After installing the extension, the side activity bar will display an icon for **Cody**. Click this icon, and Cody's `Chat` panel will open. This interface is used to ask Cody questions and paste in code snippets.

Cody also supports `Commands` with VS Code. These are quick, ready-to-use prompt actions that you can apply to any code or text-based snippet you've highlighted. You can run a command in 3 ways:

1. Type `/` in the chat bar, and Cody will suggest a list of commands
2. Right click -> Cody -> Select a command
3. Press the command hotkey (`⌥` + `c` / `alt` + `c`)

## Working with Cody extension

Let's write a JavaScript function that converts a `given date` into a human-readable description of the time elapsed between the `given date` and the `current date`. Next, you'll use this code to perform the following tasks with Cody:

- Try the `Generate Unit Tests` command
- Ask Cody to pull information from the documentation
- Ask Cody to write context-aware code

Here's the code for the date elapsed function from `date.js` file:

```js
function getTimeAgoDescription(dateString) {
	const startDate = new Date(dateString);
	const currentDate = new Date();

	const years = currentDate.getFullYear() - startDate.getFullYear();
	const months = currentDate.getMonth() - startDate.getMonth();
	const days = currentDate.getDate() - startDate.getDate();

	let timeAgoDescription = '';

	if (years > 0) {
		timeAgoDescription += `${years} ${years === 1 ? 'year' : 'years'}`;
	}

	if (months > 0) {
		if (timeAgoDescription !== '') {
			timeAgoDescription += ' ';
		}
		timeAgoDescription += `${months} ${months === 1 ? 'month' : 'months'}`;
	}

	if (days > 0) {
		if (timeAgoDescription !== '') {
			timeAgoDescription += ' ';
		}
		timeAgoDescription += `${days} ${days === 1 ? 'day' : 'days'}`;
	}

	if (timeAgoDescription === '') {
		timeAgoDescription = 'today';
	} else {
		timeAgoDescription += ' ago';
	}

	return timeAgoDescription;
}

const inputDate = '2020-02-09';
const description = getTimeAgoDescription(inputDate);
console.log(description);

const outputElement = document.getElementById('output');
outputElement.textContent = description;
```

## Generate a unit test

To ensure code quality and early bug detection, one of the most useful commands that Cody offers is `Generate Unit Tests`. It quickly helps you write a test code for any snippet that you have highlighted. To generate a unit test for our example function:

- Open the `date.js` file in VS Code
- Highlight a code snippet that you'd like to test
- Inside Cody chat type `/test` or press the command hotkey (`⌥` + `c` / `alt` + `c`)
- Select `Generate Unit Tests` option and hit `Enter`

Cody will help you generate the following unit test in the sidebar:

```js
import assert from 'assert';
import { getTimeAgoDescription } from '../src/date.js';

describe('getTimeAgoDescription', () => {
	it('returns correct relative time for date 1 month ago', () => {
		const oneMonthAgo = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
		const result = getTimeAgoDescription(oneMonthAgo);
		assert.equal(result, '1 month ago');
	});

	it('returns correct relative time for date 2 months ago', () => {
		const twoMonthsAgo = new Date(Date.now() - 2 * 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
		const result = getTimeAgoDescription(twoMonthsAgo);
		assert.equal(result, '2 months ago');
	});
});

```

This unit test file tests the `getTimeAgoDescription()` function from the `date.js` file. These tests help you validate that the `getTimeAgoDescription()` function correctly returns the relative time descriptions for dates in the past. The tests generate specific sample input dates and confirms that the output matches the expected result.

## Suggest bug fixes and changes to code snippets

Cody is very efficient at catching bugs and suggesting improvements to code snippets. To try this out, let's run our previously generated unit test and see if Cody notices any issues. Inside your VS Code terminal run the following command to try the unit test:

```bash
node tests/test.js
```

This results in an error, as our function does not correctly handle dates more than a year ago.

![Example of running failed unit test ](https://storage.googleapis.com/sourcegraph-assets/Docs/cody-quickstart/unit-test-fail.png)

Let's paste this error into the Cody chat and see what suggestions it provides:

![Example of error debugging with Cody ](https://storage.googleapis.com/sourcegraph-assets/Docs/cody-quickstart/debug-with-cody.png)

Leveraging the failed test output, Cody is able to identify the potential bug and suggest a fix. It recommends to install `mocha`, importing it at the top of the `test.js` file to identify `describe` and finally running it with `mocha`.

![Example of running failed unit test ](https://storage.googleapis.com/sourcegraph-assets/Docs/cody-quickstart/passed-tests.png)

## Ask Cody to pull reference documentation

Cody can also directly reference documentation. If Cody has context of a codebase, and docs are committed within that codebase, Cody can search through the text of docs to understand documentation and quickly pull out information so that you don't have to search for it yourself.

In the `Chat` tab of the VS Code extension, you can ask Cody examples like:

- "What happens when Service X runs out of memory?"
- "Where can I find a list of all experimental features and their feature flags?"
- "How is the changelog generated?"

## Ask Cody to write context-aware code

One of Cody's strengths is writing code, especially boilerplate code or simple code that requires general awareness of the broader codebase.

One great example of this is writing an API call. Once Cody has context of a codebase, including existing API schemas within that codebase, you can ask Cody to write code for API calls.

For example, ask Cody:

- "Write an API call to retrieve user permissions data"

## Try other commands and Cody chat

Cody can do many other things, including:

1. Explain code snippets
2. Refactor code and variable names
3. Translate code to different languages
4. Answer general questions about your code
5. Suggest bugfixes and changes to code snippets

[cody-with-sourcegraph]: overview//cody-with-sourcegraph.md
