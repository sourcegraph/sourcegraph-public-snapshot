import { describe, expect, it } from '@jest/globals'

import { flattenTree } from './Tree'

describe('Tree', () => {
    describe('flattenTree', () => {
        it('creates a flat list of TreeNode[]', () => {
            const tree = {
                name: 'root',
                children: [
                    {
                        name: 'child1',
                        children: [
                            {
                                name: 'child1.1',
                                children: [],
                            },
                            {
                                name: 'child1.2',
                                children: [],
                            },
                        ],
                    },
                    {
                        name: 'child2',
                    },
                ],
            }

            const flatTree = flattenTree(tree)
            expect(flatTree).toEqual([
                {
                    children: [1, 4],
                    id: 0,
                    name: 'root',
                    parent: null,
                },
                {
                    children: [2, 3],
                    id: 1,
                    name: 'child1',
                    parent: 0,
                },
                {
                    children: [],
                    id: 2,
                    name: 'child1.1',
                    parent: 1,
                },
                {
                    children: [],
                    id: 3,
                    name: 'child1.2',
                    parent: 1,
                },
                {
                    children: [],
                    id: 4,
                    name: 'child2',
                    parent: 0,
                },
            ])
        })

        it('passes through all properties of the node', () => {
            const tree = {
                name: 'root',
                mySuperSecretNode: "I'm a secret!",
                children: [
                    {
                        name: 'child1',
                        whoAmI: 'I am child1',
                    },
                ],
            }

            const flatTree = flattenTree(tree)
            expect(flatTree).toEqual([
                {
                    children: [1],
                    id: 0,
                    name: 'root',
                    mySuperSecretNode: "I'm a secret!",
                    parent: null,
                },
                {
                    children: [],
                    id: 1,
                    name: 'child1',
                    whoAmI: 'I am child1',
                    parent: 0,
                },
            ])
        })
    })
})
