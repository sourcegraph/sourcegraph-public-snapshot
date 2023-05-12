#!/usr/bin/env python3

import argparse
import asyncio
import collections
import math
import os
import random
import requests
import sys

from concurrent.futures import ThreadPoolExecutor

import hjson  # pip install hjson


anthropic_key = os.getenv('ANTHROPIC_KEY')
if not anthropic_key:
    print('set ANTHROPIC_KEY and re-run')
    sys.exit(1)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("examples")
    args = parser.parse_args()
    with open(args.examples) as example_file:
        data = hjson.load(example_file)

    prompt = data['seed_prompt']
    asyncio.get_event_loop().run_until_complete(churn(data, prompt))


async def churn(data, prompt):
    # Pick N examples at random.
    N = 3
    examples = random.sample(data['examples'], k=N)

    completions = apply_parallel_prompts([(prompt, example['input']) for example in examples])

    for completion in asyncio.as_completed(completions):
        print(await completion)


def apply_parallel_prompts(arg_bundles):
    with ThreadPoolExecutor(max_workers=4) as executor:
        loop = asyncio.get_event_loop()
        return [loop.run_in_executor(executor, apply_prompt, *args) for args in arg_bundles]


# TODO: Give the optimizer access to metaparameters like temperature.
def apply_prompt(prompt: str, inp):
    for k, v in inp.items():
        # TODO: don't recursively replace things here.
        prompt = prompt.replace(f'{{{k}}}', v)
    return submit_prompt(prompt)['completion']


def submit_prompt(prompt):
    """Submit a prompt to the Anthropic language model API."""
    url = "https://api.anthropic.com/v1/complete"
    headers = {
        "x-api-key": anthropic_key,
        "content-type": "application/json"
    }
    # TODO: Give the system access to the hyperparameters like temperature.
    data = {
        "prompt": prompt,
        "model": "claude-v1.3",
        "max_tokens_to_sample": 2000,
        "temperature": 0.5,
        "top_k": 3,
        "stop_sequences": ["\n\nHuman: "]
    }
    response = requests.post(url, headers=headers, json=data)
    response.raise_for_status()
    return response.json()

def explore():
    problem_framing = '''We are developing a prompt for a coding task. The output from the LLM should be written in <cody-replace> tags, repeat the included code verbatim, and address the instruction. The instruction is around the "selected" point in the code.

Input 1:

{instruction: 'document what does this program does',
 filename: 'hello.rs',
 prefix: '',
 selected: `
fn main() {
    println("Hello, world! From Cody demo land.")
}`,
 suffix: ''
}

Output 1:

<cody-replace>
// Prints a greeting to the console.
fn main() {
    println("Hello, world! From Cody demo land.")
}
</cody-replace>

Input 2:

{instruction: 'write a doc comment explaining what this function does',
 filename: 'main.rs',
 prefix: `    if denominator != 0 && numerator % denominator == 0 {
        numerator / denominator
    } else {
        i32::MIN
    }
}

`,
  selected: `fn digit(n: u16) -> &'static str {
    match n {
        0b10 => ">1",
        0b100 => ">2",
        0b1000 => ">3",
        0b10000 => ">4",
        0b100000 => ">5",
        0b1000000 => ">6",
        0b10000000 => ">7",
        0b100000000 => ">8",
        0b1000000000 => ">9",
        _ if n.count_ones() == 0 => "!",
        _ if n.count_ones() == 1 => "?1",
        _ if n.count_ones() == 2 => "?2",
        _ if n.count_ones() == 3 => "?3",
        _ if n.count_ones() == 4 => "?4",
        _ if n.count_ones() == 5 => "?5",
        _ if n.count_ones() == 6 => "?6",
        _ if n.count_ones() == 7 => "?7",
        _ if n.count_ones() == 8 => "?8",
        _ if n.count_ones() == 9 => "?9",
        _ => panic!("unreachable"),
    }
}

`,
  suffix: `const DIM: usize = 9;
#[derive(Clone)]
struct Board {
    cells: [u16; DIM * DIM],
}

`}

Output 2:

<cody-replace>    if denominator != 0 && numerator % denominator == 0 {
        numerator / denominator
    } else {
        i32::MIN
    }
}

/// Converts a number to a string indicating which bits are set.
/// If the N-th bit is set, 1 <= N < 10, the representation is
/// ">N". If no bits are set, the representation is "!". Otherwise,
/// if M bits are set, 1 <= M < 10 the output is "?M". Panics if
/// more than 9 bits are set.
fn digit(n: u16) -> &'static str {
    match n {
        0b10 => ">1",
        0b100 => ">2",
        0b1000 => ">3",
        0b10000 => ">4",
        0b100000 => ">5",
        0b1000000 => ">6",
        0b10000000 => ">7",
        0b100000000 => ">8",
        0b1000000000 => ">9",
        _ if n.count_ones() == 0 => "!",
        _ if n.count_ones() == 1 => "?1",
        _ if n.count_ones() == 2 => "?2",
        _ if n.count_ones() == 3 => "?3",
        _ if n.count_ones() == 4 => "?4",
        _ if n.count_ones() == 5 => "?5",
        _ if n.count_ones() == 6 => "?6",
        _ if n.count_ones() == 7 => "?7",
        _ if n.count_ones() == 8 => "?8",
        _ if n.count_ones() == 9 => "?9",
        _ => panic!("unreachable"),
    }
}

const DIM: usize = 9;
#[derive(Clone)]
struct Board {
    cells: [u16; DIM * DIM],
}

</cody-replace>

Input 3:

{instruction: `write a react component with a pair of color pickers
we want one of the colors to always be the complementary color of the other`,
 filename: 'scheme.tsx',
 prefix: '',
 selected: '',
 suffix: ''
}

Output 3:

<cody-replace>
import React, { Component } from 'react';
import { ChromePicker } from 'react-color';
import Color from 'color';

class ColorPicker extends Component {
    state = {
        color: '#ff0000',
        compColor: '#00ffff',
    };

    handleColorChange = (color) => {
        this.setState({
            color: color.hex,
            compColor: Color(color.hex).rotate(180).hex(),
        });
    };

    handleCompColorChange = (color) => {
        this.setState({
            color: Color(color.hex).rotate(180).hex(),
            compColor: color.hex,
        });
    };

    render() {
        return (
            <div>
                <ChromePicker
                    color={this.state.color}
                    onChangeComplete={this.handleColorChange}
                />
                <ChromePicker
                    color={this.state.compColor}
                    onChangeComplete={this.handleCompColorChange}
                />
            </div>
        );
    }
}

export default ColorPicker;
</cody-replace>

Input 4:

{instruction: 'Rewrite this to use input type="color" and your own color calculation, not those external packages',
 filename: 'color-picker.js',
 prefix: '',
 selection: '',
 suffix: `import React, { Component } from 'react';
import { ChromePicker } from 'react-color';
import Color from 'color';

class ColorPicker extends Component {
    state = {
        color: '#ff0000',
        compColor: '#00ffff',
    };

    handleColorChange = (color) => {
        this.setState({
            color: color.hex,
            compColor: Color(color.hex).rotate(180).hex(),
        });
    };

    handleCompColorChange = (color) => {
        this.setState({
            color: Color(color.hex).rotate(180).hex(),
            compColor: color.hex,
        });
    };

    render() {
        return (
            <div>
                <ChromePicker
                    color={this.state.color}
                    onChangeComplete={this.handleColorChange}
                />
                <ChromePicker
                    color={this.state.compColor}
                    onChangeComplete={this.handleCompColorChange}
                />
            </div>
        );
    }
}

export default ColorPicker;
`
}

Output 4:

<cody-replace>
import React, { Component } from 'react';

class ColorPicker extends Component {
    state = {
        color: '#ff0000',
        compColor: '#00ffff',
    };

    handleColorChange = (event) => {
        this.setState({
            color: event.target.value,
            compColor: this.getComplementaryColor(event.target.value),
        });
    };

    handleCompColorChange = (event) => {
        this.setState({
            color: this.getComplementaryColor(event.target.value),
            compColor: event.target.value,
        });
    };

    getComplementaryColor = (hexColor) => {
        const rgbColor = this.hexToRgb(hexColor);
        const compRgb = { r: 255 - rgbColor.r, g: 255 - rgbColor.g, b: 255 - rgbColor.b };
        return this.rgbToHex(compRgb.r, compRgb.g, compRgb.b);
    };

    hexToRgb = (hex) => {
        const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
        return result
            ? {
                r: parseInt(result[1], 16),
                g: parseInt(result[2], 16),
                b: parseInt(result[3], 16),
              }
            : null;
    };

    rgbToHex = (r, g, b) => {
        return "#" + ((1 << 24) | ((r & 255) << 16) | ((g & 255) << 8) | (b & 255)).toString(16).slice(1).toUpperCase();
    };

    render() {
        return (
            <div>
                <input
                    type="color"
                    value={this.state.color}
                    onChange={this.handleColorChange}
                />
                <input
                    type="color"
                    value={this.state.compColor}
                    onChange={this.handleCompColorChange}
                />
            </div>
        );
    }
}

export default ColorPicker;
</cody-replace>
'''

    step_1_prompt = f'''Human: You are a machine learning expert working on writing a prompt for a large language model (LLM).

Given the following representative pieces of input as JSON and outputs as text, write a prompt suitable for a large language model. Keep in mind that the prompts should be structured and clear enough to guide the model to produce desired responses. Here are the inputs and outputs:

{problem_framing}

Please use these examples to guide the creation of new prompts for the language model. The prompt should have placeholders for the specific inputs.

Assistant: Let's work this out in a step by step way to be sure we have the right answer. '''

    #step_1 = submit_prompt(step_1_prompt)
    #print(step_1['completion'])

    step_1_llm_outputs = '''

*** Response Option 1: ***

1. We are creating prompts for a language model to generate code.
2. The prompts should have placeholders for the specific inputs.
3. The outputs should be wrapped in <cody-replace> tags.
4. The outputs should repeat the included code verbatim.
5. The outputs should address the instruction.

Based on this, here is a suggested prompt:

Given the input:
{instruction: '<instruction>',
 filename: '<filename>',
 prefix: '<prefix>',
 selected: '<selected>',
 suffix: '<suffix>'
}

Generate the output code:
<cody-replace>
<prefix>
<selected>
</cody-replace>
<suffix>

This prompt has placeholders for the instruction, filename, prefix code, selected code, and suffix code. The output will repeat the prefix, selected, and suffix code verbatim and wrap the selected code in <cody-replace> tags to indicate that is the part the model should generate to address the instruction.

Does this look like an appropriate prompt structure for the examples given? Let me know if you have any feedback or suggestions for improvement. I'm happy to revise this.

*** Response Option 2: ***

1. The prompt should have clear instructions for the LLM to generate the desired output.  Based on the examples, the instructions should specify:
- To generate output within <cody-replace> tags
- To repeat the included code verbatim
- To address the specific instruction provided in the input

2. The prompt should have placeholders for the filename, prefix code, selected code, suffix code, and instruction. Something like this:

<prompt>
For the {filename} code sample:

{prefix}
{selected}
{suffix}

{instruction}

<cody-replace>
</cody-replace>
</prompt>

3. Filling in the placeholders with the first input example gives us:

<prompt>
For the hello.rs code sample:

fn main() {
    println("Hello, world! From Cody demo land.")
}

document what does this program does

<cody-replace>
</cody-replace>
</prompt>

4. The output for the first input example fills in the <cody-replace> tags:

<prompt>
For the hello.rs code sample:

fn main() {
    println("Hello, world! From Cody demo land.")
}

document what does this program does

<cody-replace>
// Prints a greeting to the console.
fn main() {
    println("Hello, world! From Cody demo land.")
}
</cody-replace>
</prompt>

5. Repeating the same process for the remaining examples gives us the full prompt:

<prompt>
For the {filename} code sample:

{prefix}
{selected}
{suffix}

{instruction}

<cody-replace>
</cody-replace>
</prompt>

Does this make sense? Let me know if you have any other questions!

*** Response Option 3: ***

 Here is a first attempt at a prompt:

To generate a response for a coding task:

<prompt>
Here is a {filename} with the following instruction:
{instruction}

{prefix}
<cody-replace>
{selected}
</cody-replace>
{suffix}
</prompt>

For the first input, the prompt with placeholders filled in would be:

<prompt>
Here is a hello.rs with the following instruction:
document what does this program does

fn main() {
    println("Hello, world! From Cody demo land.")
}

<cody-replace>
// Prints a greeting to the console.
fn main() {
    println("Hello, world! From Cody demo land.")
}
</cody-replace>
</prompt>

For the second input, the prompt would be:

<prompt>
Here is a main.rs with the following instruction:
write a doc comment explaining what this function does

fn digit(n: u16) -> &'static str {
    match n {
        0b10 => ">1",
        0b100 => ">2",
        0b1000 => ">3",
        0b10000 => ">4",
        0b100000 => ">5",
        0b1000000 => ">6",
        0b10000000 => ">7",
        0b100000000 => ">8",
        0b1000000000 => ">9",
        _ if n.count_ones() == 0 => "!",
        _ if n.count_ones() == 1 => "?1",
        _ if n.count_ones() == 2 => "?2",
        _ if n.count_ones() == 3 => "?3",
        _ if n.count_ones() == 4 => "?4",
        _ if n.count_ones() == 5 => "?5",
        _ if n.count_ones() == 6 => "?6",
        _ if n.count_ones() == 7 => "?7",
        _ if n.count_ones() == 8 => "?8",
        _ if n.count_ones() == 9 => "?9",
        _ => panic!("unreachable"),
    }
}

<cody-replace>
/// Converts a number to a string indicating which bits are set.
/// If the N-th bit is set, 1 <= N < 10, the representation is
/// ">N". If no bits are set, the representation is "!". Otherwise,
/// if M bits are set, 1 <= M < 10 the output is "?M". Panics if
/// more than 9 bits are set.
fn digit(n: u16) -> &'static str {
    match n {
        0b10 => ">1",
        0b100 => ">2",
        0b1000 => ">3",
        0b10000 => ">4",
        0b100000 => ">5",
        0b1000000 => ">6",
        0b10000000 => ">7",
        0b100000000 => ">8",
        0b1000000000 => ">9",
        _ if n.count_ones() == 0 => "!",
        _ if n.count_ones() == 1 => "?1",
        _ if n.count_ones() == 2 => "?2",
        _ if n.count_ones() == 3 => "?3",
        _ if n.count_ones() == 4 => "?4",
        _ if n.count_ones() == 5 => "?5",
        _ if n.count_ones() == 6 => "?6",
        _ if n.count_ones() == 7 => "?7",
        _ if n.count_ones() == 8 => "?8",
        _ if n.count_ones() == 9 => "?9",
        _ => panic!("unreachable"),
    }
}
</cody-replace>
</prompt>

*** Response Option 4: ***

The prompts should be structured and clear enough to guide the model to produce desired responses. Here are the inputs and output
s:

You are writing a prompt for a coding task. The output from the LLM should be written in <cody-replace> tags, repeat the included
 code verbatim, and address the instruction. The instruction is around the "selected" point in the code.

Input 1:

{instruction: 'document what does this program does',
 filename: 'hello.rs',
 prefix: '',
 selected: `
fn main() {
    println("Hello, world! From Cody demo land.")
}`,
 suffix: ''
}

Output 1:

<cody-replace>
// Prints a greeting to the console.
fn main() {
    println("Hello, world! From Cody demo land.")
}
</cody-replace>

Prompt: Here is a code sample with {selected} selected. Please {instruction} and wrap your response in <cody-replace> tags:

{prefix}
{selected}
{suffix}

Input 2:

{instruction: 'write a doc comment explaining what this function does',
 filename: 'main.rs',
 prefix: `    if denominator != 0 && numerator % denominator == 0 {
        numerator / denominator
    } else {
        i32::MIN
    }
}

`,
  selected: `fn digit(n: u16) -> &'static str {
    match n {
        0b10 => ">1",
        0b100 => ">2",
        0b1000 => ">3",
        0b10000 => ">4",
        0b100000 => ">5",
        0b1000000 => ">6",
        0b10000000 => ">7",
        0b100000000 => ">8",
        0b1000000000 => ">9",
        _ if n.count_ones() == 0 => "!",
        _ if n.count_ones() == 1 => "?1",
        _ if n.count_ones() == 2 => "?2",
        _ if n.count_ones() == 3 => "?3",
        _ if n.count_ones() == 4 => "?4",
        _ if n.count_ones() == 5 => "?5",
        _ if n.count_ones() == 6 => "?6",
        _ if n.count_ones() == 7 => "?7",
        _ if n.count_ones() == 8 => "?8",
        _ if n.count_ones() == 9 => "?9",
        _ => panic!("unreachable"),
    }
}

`,
  suffix: `const DIM: usize = 9;
#[derive(Clone)]
struct Board {
    cells: [u16; DIM * DIM],
}

`}

Output 2:

<cody-replace>    if denominator != 0 && numerator % denominator == 0 {
        numerator / denominator
    } else {
        i32::MIN
    }
}

/// Converts a number to a string indicating which bits are set.
/// If the N-th bit is set, 1 <= N < 10, the representation is
/// ">N". If no bits are set, the representation is "!". Otherwise,
/// if M bits are set, 1 <= M < 10 the output is "?M". Panics if
/// more than 9 bits are set.
fn digit(n: u16) -> &'static str {
    match n {
        0b10 => ">1",
        0b100 => ">2",
        0b1000 => ">3",
        0b10000 => ">4",
        0b100000 => ">5",
        0b1000000 => ">6",
        0b10000000 => ">7",
        0b100000000 => ">8",
        0b1000000000 => ">9",
        _ if n.count_ones() == 0 => "!",
        _ if n.count_ones() == 1 => "?1",
        _ if n.count_ones() == 2 => "?2",
        _ if n.count_ones() == 3 => "?3",
        _ if n.count_ones() == 4 => "?4",
        _ if n.count_ones() == 5 => "?5",
        _ if n.count_ones() == 6 => "?6",
        _ if n.count_ones() == 7 => "?7",
        _ if n.count_ones() == 8 => "?8",
        _ if n.count_ones() == 9 => "?9",
        _ => panic!("unreachable"),
    }
}

const DIM: usize = 9;
#[derive(Clone)]
struct Board {
    cells: [u16; DIM * DIM],
}

</cody-replace>

Prompt: Here is a code sample with {selected} selected. Please {instruction} and wrap your response in <cody-replace> tags:

{prefix}
{selected}
{suffix}
'''

    step_2_prompt = f'''Human: {problem_framing}

You are a researcher tasked with investigating the 4 response options provided. List the flaws and and faulty logic of each answer option.

{step_1_llm_outputs}

Assistant: Let's work this out in a step by step way to be sure we have all the errors: '''

    #step_2 = submit_prompt(step_2_prompt)
    #print(step_2['completion'])

    step_2_llm_output = '''
1. The prompt should have clear instructions for the LLM to generate the desired output.
- The prompt should specify to generate output within <cody-replace> tags
- The prompt should specify to repeat the included code verbatim
- The prompt should specify to address the specific instruction provided in the input

2. The prompt should have placeholders for the filename, prefix code, selected code, suffix code, and instruction. Something like this:

<prompt>
For the {filename} code sample:

{prefix}
{selected}
{suffix}

{instruction}

<cody-replace>
</cody-replace>
</prompt>

3. Filling in the placeholders with the first input example gives us:

<prompt>
For the hello.rs code sample:

fn main() {
    println("Hello, world! From Cody demo land.")
}

document what does this program does

<cody-replace>
</cody-replace>
</prompt>

4. The output for the first input example fills in the <cody-replace> tags:

<prompt>
For the hello.rs code sample:

fn main() {
    println("Hello, world! From Cody demo land.")
}

document what does this program does

<cody-replace>
// Prints a greeting to the console.
fn main() {
    println("Hello, world! From Cody demo land.")
}
</cody-replace>
</prompt>

5. Repeating the same process for the remaining examples gives us the full prompt:

<prompt>
For the {filename} code sample:

{prefix}
{selected}
{suffix}

{instruction}

<cody-replace>
</cody-replace>
</prompt>

Flaws in Response Option 1:
- The prompt does not specify clearly enough to generate output within <cody-replace> tags. It only has placeholders for the tags but does not instruct the model to generate output within them.
- The prompt does not specify clearly enough to repeat the included code verbatim. It only has placeholders for the code but does not instruct the model to repeat it.
- The prompt does not specify clearly enough to address the specific instruction. It only has a placeholder for the instruction but does not instruct the model to address it.

Flaws in Response Option 2:
- No major flaws identified. This prompt addresses the specifications clearly.

Flaws in Response Option 3:
- The prompt does not specify clearly enough to generate output within <cody-replace> tags. It only has placeholders for the tags but does not instruct the model to generate output within them.
- The prompt does not specify clearly enough to repeat the included code verbatim. It only has placeholders for the code but does not instruct the model to repeat it.
- The prompt does specify clearly to address the specific instruction.

Flaws in Response Option 4:
- No major flaws identified. This prompt addresses the specifications clearly.

In summary, Response Options 2 and 4 provide the clearest prompts by specifying all required instructions for the model. Response Options 1 and 3 lack clarity in instructing the model to generate output in the tags and repeat the code. Please let me know if you have any other questions!
'''

    step_3_prompt = f'''Human: {problem_framing}

Here are 4 response options:

{step_1_llm_outputs}

Researcher: {step_2_llm_output}

You are a resolver tasked with 1) finding which of the 4 prompt options the researcher thought was best 2) improving that prompt, and 3) printing the improved prompt in full.

Assistant: Let's work this out in a step by step way to be sure we have the right answer. '''

    #step_3 = submit_prompt(step_3_prompt)
    #print(step_3['completion'])

    step_3_llm_output = '''
1) The researcher thought Response Option 2 was the best prompt. They said "Response Options 2 and 4 provide the clearest prompts by specifying all required instructions for the model." However, they did not identify any flaws in Response Option 2, whereas they did find a minor flaw in Response Option 4. So Response Option 2 seems to be their preferred choice.
2) To improve the prompt, I would:
- Add more context about the purpose of the task, e.g. "You are generating code samples to demonstrate your abilities."
- Specify the language the code is in, e.g. "The code samples are in Rust."
- Add formatting to make the prompt visually clear, e.g. section headers, indentation, etc.

Here is the improved prompt:

You are generating code samples to demonstrate your abilities. The code samples are in Rust.

For the {filename} code sample:

{prefix}
{selected}
{suffix}

{instruction}

<cody-replace>
</cody-replace>
</prompt>

3) Here is the full improved prompt:

You are generating code samples to demonstrate your abilities. The code samples are in Rust.

For the {filename} code sample:

{prefix}
{selected}
{suffix}

{instruction}

<cody-replace>
</cody-replace>
</prompt>

Please let me know if you would like me to explain or expand on any part of this response. I'm happy to provide more details and clarification.
'''

    trial_1_desired_output = '''<cody-replace>
import React, { Component } from 'react';

class ColorPicker extends Component {
    state = {
        color: '#ff0000',
        compColor: '#00ffff',
    };

    handleColorChange = (event) => {
        this.setState({
            color: event.target.value,
            compColor: this.getComplementaryColor(event.target.value),
        });
    };

    handleCompColorChange = (event) => {
        this.setState({
            color: this.getComplementaryColor(event.target.value),
            compColor: event.target.value,
        });
    };

    getComplementaryColor = (hexColor) => {
        const rgbColor = this.hexToRgb(hexColor);
        const compRgb = { r: 255 - rgbColor.r, g: 255 - rgbColor.g, b: 255 - rgbColor.b };
        return this.rgbToHex(compRgb.r, compRgb.g, compRgb.b);
    };

    hexToRgb = (hex) => {
        const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
        return result
            ? {
                r: parseInt(result[1], 16),
                g: parseInt(result[2], 16),
                b: parseInt(result[3], 16),
              }
            : null;
    };

    rgbToHex = (r, g, b) => {
        return "#" + ((1 << 24) | ((r & 255) << 16) | ((g & 255) << 8) | (b & 255)).toString(16).slice(1).toUpperCase();
    };

    render() {
        return (
            <div>
                <input
                    type="color"
                    value={this.state.color}
                    onChange={this.handleColorChange}
                />
                <input
                    type="color"
                    value={this.state.compColor}
                    onChange={this.handleCompColorChange}
                />
            </div>
        );
    }
}

export default ColorPicker;
</cody-replace>
'''

    trial_1_prompt = '''You are generating code samples to demonstrate your abilities. The code samples are in Rust.

For the color-picker.js code sample:

import React, { Component } from 'react';
import { ChromePicker } from 'react-color';
import Color from 'color';

class ColorPicker extends Component {
    state = {
        color: '#ff0000',
        compColor: '#00ffff',
    };

    handleColorChange = (color) => {
        this.setState({
            color: color.hex,
            compColor: Color(color.hex).rotate(180).hex(),
        });
    };

    handleCompColorChange = (color) => {
        this.setState({
            color: Color(color.hex).rotate(180).hex(),
            compColor: color.hex,
        });
    };

    render() {
        return (
            <div>
                <ChromePicker
                    color={this.state.color}
                    onChangeComplete={this.handleColorChange}
                />
                <ChromePicker
                    color={this.state.compColor}
                    onChangeComplete={this.handleCompColorChange}
                />
            </div>
        );
    }
}

export default ColorPicker;

Rewrite this to use input type="color" and your own color calculation, not those external packages

<cody-replace>
</cody-replace>
</prompt>'''

    # TODO: Instantiate the template
    trial_1_inputs = '''{instruction: 'Rewrite this to use input type="color" and your own color calculation, not those external packages',
 filename: 'color-picker.js',
 prefix: '',
 selection: '',
 suffix: `import React, { Component } from 'react';
import { ChromePicker } from 'react-color';
import Color from 'color';

class ColorPicker extends Component {
    state = {
        color: '#ff0000',
        compColor: '#00ffff',
    };

    handleColorChange = (color) => {
        this.setState({
            color: color.hex,
            compColor: Color(color.hex).rotate(180).hex(),
        });
    };

    handleCompColorChange = (color) => {
        this.setState({
            color: Color(color.hex).rotate(180).hex(),
            compColor: color.hex,
        });
    };

    render() {
        return (
            <div>
                <ChromePicker
                    color={this.state.color}
                    onChangeComplete={this.handleColorChange}
                />
                <ChromePicker
                    color={this.state.compColor}
                    onChangeComplete={this.handleCompColorChange}
                />
            </div>
        );
    }
}

export default ColorPicker;
`
}
'''
    trial_1_prompt_template = '''You are generating code samples to demonstrate your abilities. The code samples are in Rust.

For the {filename} code sample:

{prefix}
{selected}
{suffix}

{instruction}

<cody-replace>
</cody-replace>
</prompt>
'''

    #trial_1 = submit_prompt(trial_1_prompt)
    #print(trial_1['completion'])
    trial_1_actual = '''

Here is the rewritten code sample in Rust:

```rust
use std::cmp::Ordering;

enum Color {
    Red,
    Green,
    Blue,
}

impl Color {
    fn to_hex(&self) -> &'static str {
        match self {
            Color::Red => "#ff0000",
            Color::Green => "#00ff00",
            Color::Blue => "#0000ff",
        }
    }
}

fn complementary_color(color: &Color) -> Color {
    match color {
        Color::Red => Color::Green,
        Color::Green => Color::Red,
        Color::Blue => Color::Red,
    }
}

fn rotate_color(color: &Color, degrees: i32) -> Color {
    let mut new_color = color.clone();
    for _ in 0..degrees {
        new_color = complementary_color(&new_color);
    }
    new_color
}

fn main() {
    let mut color = Color::Red;
    let mut comp_color = Color::Green;

    println!("Color: {}", color.to_hex());
    println!("Complementary color: {}", comp_color.to_hex());

    let new_color = rotate_color(&color, 180);
    color = new_color;
    comp_color = complementary_color(&color);

    println!("Color: {}", color.to_hex());
    println!("Complementary color: {}", comp_color.to_hex());
}
```

This code sample defines an enum `Color` to represent red, green and blue. It implements methods to get the hex code for a color, get the complementary color, and rotate a color by a given number of degrees. It then demonstrates using these methods by starting with red and green, rotating red by 180 degrees, and updating the complementary green color.
'''

    critic_prompt = f'''Human: {problem_framing}

Our research team came up with this prompt template:

<prompt>
{trial_1_prompt_template}
</prompt>

We used this prompt template with these inputs:

{trial_1_inputs}

The model generated this output:

{trial_1_actual}

We wanted this output:

{trial_1_desired_output}

You are a researcher tasked with investigating the flaws in the prompt provided. List the flaws. Let's work this out in a step by step way to be sure we have all the errors.

Assistant: '''

    #critic = submit_prompt(critic_prompt)
    #print(critic['completion'])
    critic_criticism = ''' Here are the flaws I identified in the prompt:

1. The prompt specifies the code sample is in Rust, but the output is in JavaScript (React). The language should match between the prompt and output.

2. The prompt includes {prefix}, {selected}, and {suffix} fields but does not actually use them in the prompt text. These fields should either be used or removed.

3. The instruction in the prompt is "Rewrite this to use input type="color" and your own color calculation, not those external packages" but the output code is not actually a rewrite of the code in the {suffix}. The output should be a rewrite of that code to match the instruction.

4. The <cody-replace> tags in the prompt are not closed, they should be <cody-replace> </cody-replace>

5. The prompt does not specify what the model is supposed to generate within the <cody-replace> tags. It should specify "Generate the code sample here:" or something similar.

6. The output includes a code comment "// Prints a greeting to the console." which does not match the actual functionality of printing complementary colors. The comment should be updated to match the code.

7. The output code is missing a closing } for the main() function.

8. The hexToRgb() and rgbToHex() functions in the output code are not used and can be removed.

9. The output code is missing props for the input elements and a class for the overall component.

Those are the key issues I identified in the prompt and output. Please let me know if you have any other questions!
'''

    prompt_improvement_prompt = f'''Human: {problem_framing}

Our research team came up with this prompt template:

<prompt>
{trial_1_prompt_template}
</prompt>

An independent researcher conducted this research about the prompt:

{critic_criticism}

You are a resolver tasked with 1) finding which of the independent researcher's criticisms are valid 2) improving the prompt to address those problems 3) printing the improved prompt in full. Let's work this out in a step by step way to be sure we have the right answer:

Assistant: '''
    prompt_improvement = submit_prompt(prompt_improvement_prompt)
    print(prompt_improvement['completion'])

# Observations:
# - Needs help generalizing. One example is too specific.
# - Needs help generalizing, coach it about the inputs being replaced.
# - Needs help generalizing, coach it that the prompt should not contain details of specific instances of the problem.
# - Needs to raise an exception about data being missing. In this case, the language.
# - Fan out, find something in common, fan in might work well.
# - If we started with a seed prompt, then the problem becomes homogeneous.
# - Idea, give it access to metaparameters about temperature, top K.

if __name__ == '__main__':
    main()
