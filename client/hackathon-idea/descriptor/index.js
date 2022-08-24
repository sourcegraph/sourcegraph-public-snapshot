const fetch = require('node-fetch')
const registryUrl = 'https://registry.npmjs.org/'
const { Configuration, OpenAIApi } = require('openai')

const packages = require('../src/packages.json')

const fs = require('fs')

const configuration = new Configuration({
  apiKey: process.env.OPENAI_API_KEY,
})
const openai = new OpenAIApi(configuration)

;(async () => {
  const npmResolver = async packageName => {
    try {
      const response = await fetch(`${registryUrl}${packageName.toLowerCase()}`)
      const data = await response.json()

      const name = data.name
      const description = data.description
      const license = data.license
      const homepage = data.homepage

      const package_ = {
        name,
        description,
        license,
        homepage,
      }

      return package_
    } catch (error) {
      console.log(`failed to fetch pkg ${packageName}`)
      console.log(`err: ${error}`)
      return {}
    }
  }

  const results = []

  for (const package of packages) {
    const packageData = await npmResolver(package.name)

    let description = ''

    try {
      const response = await openai.createCompletion({
        model: 'text-davinci-002',
        prompt: `Project name: react\nWhat it does: A declarative, efficient, and flexible JavaScript library for building interfaces.\nLicense: MIT\nDescription: React is a popular library written in JavaScript. It uses an MIT license. It's a declarative, efficient, and flexible library for building interfaces. It is being used by millions of developers worldwide. \n\n\nProject name: ${packageData.name}\nWhat it does: ${packageData.description}\nLicense: ${packageData.license}\nDescription:`,
        temperature: 0.5,
        max_tokens: 500,
        top_p: 1,
        frequency_penalty: 0,
        presence_penalty: 0,
      })

      description = response.data.choices[0].text
    } catch (error) {
      console.log(`open ai failed ${error}`)
    }

    results.push({
      name: package.name,
      description,
    })
  }

  fs.writeFileSync('packages.json', JSON.stringify(results), 'utf8')
})()
