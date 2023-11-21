# Generate Unit Tests

<p class="subtitle">In this guide, you'll learn how to use Cody to generate unit tests for your codebase.</p>

For a robust and reliable codebase it's important to write unit test for your code. These unit tests will help you to ensure that your code works as expected and that it doesn't break in the future. While writing these tests can appear as a cumbersome task, Cody can automatically generate unit tests for your codebase.

In this guide, you will learn to generate a unit test using Cody with real-time code examples.

## Prerequisites

Before you begin, make sure you have the following:

- Access to Sourcegraph and are logged into your account
- Cody VS Code extension installed and connected to your Sourcegraph instance

> NOTE: Cody also provides extensions for JetBrains and Neovim but these are not covered in this guide.

## Generate unit tests with Cody

One of the prominent commands in Cody is **Generate unit tests** (`/test`). You just need to select a block of code and ask Cody to write the tests for you.

For this guide, let's use an example that converts any `given date` into a human-readable description of the time elapsed between the `given date` and the `current date`.
