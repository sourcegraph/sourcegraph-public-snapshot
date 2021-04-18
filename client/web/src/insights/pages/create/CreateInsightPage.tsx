import React from 'react';
import { Form } from 'reactstrap';

import { Page } from '../../../components/Page';
import { PageTitle } from '../../../components/PageTitle';

interface CreateInsightPageProps {}

export const CreateInsightPage: React.FunctionComponent<CreateInsightPageProps> = () => (
        <Page className='create-insight col-8'>
            <PageTitle title='Create new code insight'/>

            <div className='create-insight__sub-title-container'>

                <h2 className='create-insight__sub-title'>Create new code insight</h2>

                <p>
                    Search-based code insights analyse your code based on any search query.

                    <a href="https://docs.sourcegraph.com/code_monitoring/how-tos/starting_points"
                       target="_blank"
                       rel="noopener"
                       className='create-insight__doc-link'>Learn more.</a>
                </p>
            </div>

            <Form className='create-insight__form'>

                <label>
                    <span>Name</span>

                    <input
                        type="text"
                        className="form-control my-2"
                        required={true}
                        autoFocus={true}
                        spellCheck={false}
                        placeholder='Enter the unique name for your insight'
                    />

                    <span>Chose a unique for your insights</span>
                </label>

                <label>
                    <span>Title</span>

                    <input
                        type="text"
                        className="form-control my-2"
                        required={true}
                        autoFocus={true}
                        spellCheck={false}
                        placeholder='ex. Migration to React function components'
                    />

                    <span>Shown as title for your insight</span>
                </label>

                <label>
                    <span>Repositories</span>

                    <input
                        type="text"
                        className="form-control my-2"
                        required={true}
                        autoFocus={true}
                        spellCheck={false}
                        placeholder='Add or search for repositories'
                    />

                    <span>Create a list of repositories to run your search over. Separate them with comas.</span>
                </label>

                <div>
                    <span>Visibility</span>

                    <label>
                        <input
                            type="radio"
                            className="form-control my-2"
                            required={true}
                        />

                        <div>
                            <span>Personal</span>
                            <span>Only for you</span>
                        </div>
                    </label>

                    <label>
                        <input
                            type="radio"
                            className="form-control my-2"
                            required={true}
                        />

                        <div>
                            <span>Organization</span>
                            <span>to all users in your organization</span>
                        </div>
                    </label>

                    <span>
                        This insigh will be visible only on your personal dashboard. It will not be show to other
                        users in your organisation.
                    </span>
                </div>

                <hr/>

                <div>
                    <h3>Data series</h3>

                    <div>

                        <label>
                            <span>Name</span>

                            <input
                                type="text"
                                className="form-control my-2"
                                required={true}
                                autoFocus={true}
                                spellCheck={false}
                                placeholder='ex. Function component'
                            />

                            <span>Name shown in the legend and tooltip</span>
                        </label>

                        <label>
                            <span>Query</span>

                            <input
                                type="text"
                                className="form-control my-2"
                                placeholder='ex. spatternType:regexp const\\s\\w+:\\s(React\\.)?FunctionComponent'
                            />

                            <span>Do not include the repo: filter as it will be added automatically for the current repository</span>
                        </label>

                        <div>
                            <span>Color</span>

                            <div>

                                <label>
                                    <span>Red</span>
                                    <input
                                        type="radio"
                                        className="form-control my-2"
                                        placeholder='ex. spatternType:regexp const\\s\\w+:\\s(React\\.)?FunctionComponent'
                                    />
                                </label>

                                <label>
                                    <span>Blue</span>
                                    <input
                                        type="radio"
                                        className="form-control my-2"
                                        placeholder='ex. spatternType:regexp const\\s\\w+:\\s(React\\.)?FunctionComponent'
                                    />
                                </label>

                                <label>
                                    <span>Green</span>
                                    <input
                                        type="radio"
                                        className="form-control my-2"
                                        placeholder='ex. spatternType:regexp const\\s\\w+:\\s(React\\.)?FunctionComponent'
                                    />
                                </label>
                            </div>

                            <span>Or <a href="">use custom color</a></span>
                        </div>

                        <button>Done</button>
                    </div>
                </div>

                <hr/>

                <div>
                    <h3>Step between data points</h3>

                    <div>
                        <input
                            type="text"
                            className="form-control my-2"
                            placeholder='ex. 2'
                        />

                        <label>
                            <span>Red</span>
                            <input
                                type="radio"
                                className="form-control my-2"
                                placeholder='ex. spatternType:regexp const\\s\\w+:\\s(React\\.)?FunctionComponent'
                            />
                        </label>

                        <label>
                            <span>Blue</span>
                            <input
                                type="radio"
                                className="form-control my-2"
                                placeholder='ex. spatternType:regexp const\\s\\w+:\\s(React\\.)?FunctionComponent'
                            />
                        </label>

                        <label>
                            <span>Green</span>
                            <input
                                type="radio"
                                className="form-control my-2"
                                placeholder='ex. spatternType:regexp const\\s\\w+:\\s(React\\.)?FunctionComponent'
                            />
                        </label>
                    </div>

                    <span>The distance between two data points on the chart</span>
                </div>

                <hr/>

                <div>
                    <button type='button' >Create code insight</button>
                    <button type='button'>Cancel</button>
                </div>
            </Form>
        </Page>
    )
