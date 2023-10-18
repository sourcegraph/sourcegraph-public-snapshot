# Usage and Limits

<p class="subtitle">Learn about all the usage policies and limits that apply to context.</p>

Cody's usage and limits policies help optimize its performance and ensure cost-effectiveness. This section provides insights into the context window size, token limits for chat, commands, and completions, and administrators' control over these settings.

## Cody context window size

Cody's context window represents the amount of contextual information provided to assist in generating code or responses. Currently, the context window size for both chat and commands is set to **7000**.

While Cody aims to provide maximum context for each prompt, there are limits to ensure efficiency. Cody shares all relevant context for chat but limits it to **12 code results** and **3 text results** (i.e., 15 tokens) to maintain performance.

## Manage context window size

Site administrators can update the maximum context window size to meet their specific requirements. In that case, default commands do not require the maximum 15 file limit for context, contributing to cost savings and improving response quality.
