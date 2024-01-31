package ipynb

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	out, err := Render(notebookString)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Print(out)
}

var notebookString = `
{
    "metadata": {
        "kernelspec": {
            "name": "python",
            "display_name": "Python (Pyodide)",
            "language": "python"
        },
        "language_info": {
            "codemirror_mode": {
                "name": "python",
                "version": 3
            },
            "file_extension": ".py",
            "mimetype": "text/x-python",
            "name": "python",
            "nbconvert_exporter": "python",
            "pygments_lexer": "ipython3",
            "version": "3.8"
        }
    },
    "nbformat_minor": 4,
    "nbformat": 4,
    "cells": [
        {
            "cell_type": "markdown",
            "source": "# Introduction to the JupyterLab and Jupyter Notebooks\n\nThis is a short introduction to two of the flagship tools created by [the Jupyter Community](https://jupyter.org).\n\n## Jupyter Notebooks 📓\n\n**Jupyter Notebooks** are a community standard for communicating and performing interactive computing. They are a document that blends computations, outputs, explanatory text, mathematics, images, and rich media representations of objects.\n\nJupyterLab is one interface used to create and interact with Jupyter Notebooks.\n\n**For an overview of Jupyter Notebooks**, see the **JupyterLab Welcome Tour** on this page, by going to Help -> Notebook Tour and following the prompts.\n\n> **See Also**: For a more in-depth tour of Jupyter Notebooks and the Classic Jupyter Notebook interface, see [the Jupyter Notebook IPython tutorial on Binder](https://mybinder.org/v2/gh/ipython/ipython-in-depth/HEAD?urlpath=tree/binder/Index.ipynb).\n\n## An example: visualizing data in the notebook ✨\n\nBelow is an example of a code cell. We'll visualize some simple data using two popular packages in Python. We'll use [NumPy](https://numpy.org/) to create some random data, and [Matplotlib](https://matplotlib.org) to visualize it.\n\nNote how the code and the results of running the code are bundled together.",
            "metadata": {}
        },
        {
            "cell_type": "code",
            "source": "from matplotlib import pyplot as plt\nimport numpy as np\n\n# Generate 100 random data points along 3 dimensions\nx, y, scale = np.random.randn(3, 100)\nfig, ax = plt.subplots()\n\n# Map each onto a scatterplot we'll create with Matplotlib\nax.scatter(x=x, y=y, c=scale, s=np.abs(scale)*500)\nax.set(title=\"Some random data, created with JupyterLab!\")\nplt.show()",
            "metadata": {
                "trusted": true,
                "scrolled": true
            },
            "outputs": [
                {
                    "output_type": "display_data",
                    "data": {
                        "text/plain": "<Figure size 640x480 with 1 Axes>"
                    },
                    "metadata": {}
                }
            ],
            "execution_count": 5
        },
        {
            "cell_type": "raw",
            "source": "% Raw Cell\n% This is a raw cell where you can write raw content like HTML or LaTeX.  \n% Below is an example of LaTeX math rendering.\n\n$$\nc = \\sqrt{a^2 + b^2}\n$$",
            "metadata": {
                "format": "application/x-latex"
            }
        },
        {
            "cell_type": "markdown",
            "source": "# Markdown Cell\n## LaTeX Equation\nYou can also include LaTeX equations directly in Markdown cells:\n\n$$\nf(x) = \\int_{-\\infty}^{\\infty} e^{-x^2} dx\n$$",
            "metadata": {}
        },
        {
            "cell_type": "code",
            "source": "# Error cells\n# The code below results in an error\n\nd = dict()\nprint(d[\"unknown key\"])",
            "metadata": {
                "trusted": true
            },
            "outputs": [
                {
                    "ename": "<class 'KeyError'>",
                    "evalue": "'unknown key'",
                    "traceback": [
                        "\u001b[0;31m---------------------------------------------------------------------------\u001b[0m",
                        "\u001b[0;31mKeyError\u001b[0m                                  Traceback (most recent call last)",
                        "Cell \u001b[0;32mIn[7], line 5\u001b[0m\n\u001b[1;32m      1\u001b[0m \u001b[38;5;66;03m# Error cells\u001b[39;00m\n\u001b[1;32m      2\u001b[0m \u001b[38;5;66;03m# The code below results in an error\u001b[39;00m\n\u001b[1;32m      4\u001b[0m d \u001b[38;5;241m=\u001b[39m \u001b[38;5;28mdict\u001b[39m()\n\u001b[0;32m----> 5\u001b[0m \u001b[38;5;28mprint\u001b[39m(\u001b[43md\u001b[49m\u001b[43m[\u001b[49m\u001b[38;5;124;43m\"\u001b[39;49m\u001b[38;5;124;43munknown key\u001b[39;49m\u001b[38;5;124;43m\"\u001b[39;49m\u001b[43m]\u001b[49m)\n",
                        "\u001b[0;31mKeyError\u001b[0m: 'unknown key'"
                    ],
                    "output_type": "error"
                }
            ],
            "execution_count": 7
        },
        {
            "cell_type": "raw",
            "source": "<h1>Hello, HTML!</h1><p>This is a short example of raw HTML.</p><ul><li>Item 1</li><li>Item 2</li><li>Item 3</li></ul>",
            "metadata": {
                "format": "text/html",
                "raw_mimetype": "text/html"
            }
        },
        {
            "cell_type": "markdown",
            "source": "## Cell attachements\nFollowing are [4 ways to embed images in your Jupyter notebook](https://medium.com/@yogeshkd/four-ways-to-embed-images-in-your-jupyter-notebook-powered-blog-2d28f6d1b6e6):",
            "metadata": {}
        },
        {
            "cell_type": "code",
            "source": "# Using code\nfrom IPython.display import Image\nImage(url=\"https://images.unsplash.com/photo-1612815292258-f4354f7f5c76?ixid=MXwxMjA3fDB8MHx0b3BpYy1mZWVkfDYwM3w2c01WalRMU2tlUXx8ZW58MHx8fA%3D%3D&ixlib=rb-1.2.1&auto=format&fit=crop&w=800&q=60\", height=300)",
            "metadata": {
                "trusted": true,
                "scrolled": true
            },
            "outputs": [
                {
                    "execution_count": 12,
                    "output_type": "execute_result",
                    "data": {
                        "text/html": "<img src=\"https://images.unsplash.com/photo-1612815292258-f4354f7f5c76?ixid=MXwxMjA3fDB8MHx0b3BpYy1mZWVkfDYwM3w2c01WalRMU2tlUXx8ZW58MHx8fA%3D%3D&ixlib=rb-1.2.1&auto=format&fit=crop&w=800&q=60\" height=\"300\"/>",
                        "text/plain": "<IPython.core.display.Image object>"
                    },
                    "metadata": {}
                }
            ],
            "execution_count": 12
        },
        {
            "cell_type": "markdown",
            "source": "Using Markdown syntax:  \n![Waterfall](https://images.unsplash.com/photo-1593322962878-a4b73deb1e39?ixid=MXwxMjA3fDB8MHx0b3BpYy1mZWVkfDc3NHw2c01WalRMU2tlUXx8ZW58MHx8fA%3D%3D&ixlib=rb-1.2.1&auto=format&fit=crop&w=800&q=60)",
            "metadata": {}
        },
        {
            "cell_type": "markdown",
            "source": "With an HTML tag:  \n<img src=\"https://images.unsplash.com/photo-1589652717406-1c69efaf1ff8?ixlib=rb-1.2.1&ixid=MXwxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHw%3D&auto=format&fit=crop&w=1650&q=80\" style=\"height:300px\" />",
            "metadata": {}
        },
        {
            "cell_type": "code",
            "source": "{\n    \"glossary\": {\n        \"title\": \"example glossary\",\n\t\t\"GlossDiv\": {\n            \"title\": \"S\",\n\t\t\t\"GlossList\": {\n                \"GlossEntry\": {\n                    \"ID\": \"SGML\",\n\t\t\t\t\t\"SortAs\": \"SGML\",\n\t\t\t\t\t\"GlossTerm\": \"Standard Generalized Markup Language\",\n\t\t\t\t\t\"Acronym\": \"SGML\",\n\t\t\t\t\t\"Abbrev\": \"ISO 8879:1986\",\n\t\t\t\t\t\"GlossDef\": {\n                        \"para\": \"A meta-markup language, used to create markup languages such as DocBook.\",\n\t\t\t\t\t\t\"GlossSeeAlso\": [\"GML\", \"XML\"]\n                    },\n\t\t\t\t\t\"GlossSee\": \"markup\"\n                }\n            }\n        }\n    }\n}",
            "metadata": {
                "trusted": true
            },
            "outputs": [
                {
                    "execution_count": 13,
                    "output_type": "execute_result",
                    "data": {
                        "text/plain": "{'glossary': {'title': 'example glossary',\n  'GlossDiv': {'title': 'S',\n   'GlossList': {'GlossEntry': {'ID': 'SGML',\n     'SortAs': 'SGML',\n     'GlossTerm': 'Standard Generalized Markup Language',\n     'Acronym': 'SGML',\n     'Abbrev': 'ISO 8879:1986',\n     'GlossDef': {'para': 'A meta-markup language, used to create markup languages such as DocBook.',\n      'GlossSeeAlso': ['GML', 'XML']},\n     'GlossSee': 'markup'}}}}}"
                    },
                    "metadata": {}
                }
            ],
            "execution_count": 13
        },
        {
            "cell_type": "code",
            "execution_count": 1,
            "metadata": {
                "trusted": true
            },
            "outputs": [
                {
                    "data": {
                        "application/json": {
                            "a": [
                                1,
                                2,
                                3,
                                4
                            ],
                            "b": {
                                "inner1": "helloworld",
                                "inner2": "foobar"
                            }
                        },
                        "text/plain": [
                            "<IPython.core.display.JSON object>"
                        ]
                    },
                    "execution_count": 1,
                    "metadata": {
                        "application/json": {
                            "expanded": false,
                            "root": "root"
                        }
                    },
                    "output_type": "execute_result"
                }
            ],
            "source": [
                "# This is a code cell that outputs application/json\n",
                "from IPython.display import JSON\n",
                "JSON({'a': [1, 2, 3, 4,], 'b': {'inner1': 'helloworld', 'inner2': 'foobar'}})"
            ]
        },
        {
            "cell_type": "code",
            "execution_count": 15,
            "metadata": {},
            "outputs": [
                {
                    "name": "stdout",
                    "output_type": "stream",
                    "text": [
                        "Do you approve of the following input? Anything except 'Y'/'Yes' (case-insensitive) will be treated as a no.\n",
                        "\n",
                        "ls /usr\n",
                        "yes\n",
                        "\u001b[35mX11\u001b[m\u001b[m\n",
                        "\u001b[35mX11R6\u001b[m\u001b[m\n",
                        "\u001b[1m\u001b[36mbin\u001b[m\u001b[m\n",
                        "\u001b[1m\u001b[36mlib\u001b[m\u001b[m\n",
                        "\u001b[1m\u001b[36mlibexec\u001b[m\u001b[m\n",
                        "\u001b[1m\u001b[36mlocal\u001b[m\u001b[m\n",
                        "\u001b[1m\u001b[36msbin\u001b[m\u001b[m\n",
                        "\u001b[1m\u001b[36mshare\u001b[m\u001b[m\n",
                        "\u001b[1m\u001b[36mstandalone\u001b[m\u001b[m\n",
                        "\n"
                    ]
                }
            ],
            "source": [
                "print(tool.run(\"ls /usr\"))"
            ]
        },
        {
            "cell_type": "raw",
            "source": [
                "<details><summary>Collapsible HTML</summary><strong>hi, mom!</strong></details>"
            ],
            "metadata": {
                "format": "text/html",
                "raw_mimetype": "text/html"
            }
        }
    ]
}
`
