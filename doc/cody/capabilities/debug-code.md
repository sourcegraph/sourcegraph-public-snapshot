# Debug Code

<p class="subtitle">Learn how Cody helps you identify errors in your code and provides code fixes.</p>

Cody is optimized to identify and fix errors in your code. Its debugging capability and autocomplete suggestions can significantly accelerate your debugging process, increasing the developer productivity. All Cody IDE extensions (VS Code, JetBrains, and Neovim) support code debugging and fixes capabilities.

## Use chat for code fixes

When you encounter a code error, you can use the chat interface and ask Cody about it. You can paste the faulty code snippets directly in the chat window, and Cody will provide a fix.

The suggestions can be a corrected code snippet that you can copy and paste into your code. Or a follow-up question for additional context to help you debug the code further. Moreover, you can also paste the error message in chat and ask Cody to provide you with a list of possible solutions.

Let's look at a simple example to understand how Cody can help you debug your code. The following code snippet should print the sum of two numbers.

```js
function sum(a, b) {
	var result = a + b;
	console.log('The sum is: ' + $result);
}

sum(3 , 4);
```

When you try to `console.log` the `result`, it does not print the correct summed value. Cody can help you both identify the error and provide a solution to fix it. Let's debug the code snippet. Paste the code snippet inside the Cody chat window and ask Cody to fix it.

<video width="1920" height="1080" loop playsinline controls style="width: 100%; height: auto; max-width: 50rem">
  <source src="https://storage.googleapis.com/sourcegraph-assets/Docs/Media/fix-code.mp4" type="video/mp4">
</video>

In addition, Cody helps you reduce the chances of getting syntax and typo errors. The Cody IDE extensions provide context-aware suggestions based on your codebase, helping you avoid common mistakes and reduce your debugging time.

## Detecting code smell

Cody can detect early code smells to ensure coding best practices and quality and provide suggestions to improve your code. By detecting such potential errors early on, you can avoid scalability and code maintainability issues that might arise in the future.

VS Code users can detect code smells by the `/smell` command, and JetBrains users can achieve the same by clicking the **Smell Code** button from the **Commands** tab.

Let's identify code smells on the same function from the previous example.

<video width="1920" height="1080" loop playsinline controls style="width: 100%; height: auto; max-width: 50rem">
  <source src="https://storage.googleapis.com/sourcegraph-assets/Docs/Media/code-smells.mp4" type="video/mp4">
</video>

As you can see, Cody not only detects code smells but also provides suggestions to improve your code.
