/// <reference types="cypress" />

const RepoHomepagePage = 'https://github.com/gorilla/mux'
const VIEW_FILE_BASE_BUTTON = 'View File (base)'
const VIEW_FILE_HEAD_BUTTON = 'View File (head)'

describe('GitHub', () => {
    it('.should() - assert that Sourcegraph View Repository is injected', () => {
        cy.visit(RepoHomepagePage)
        cy.get('.pagehead-actions').find('li#open-on-sourcegraph')
    })

    it('.should() - assert that BlobAnnotators are mounted for a file', () => {
        cy.visit('https://github.com/gorilla/mux/blob/master/mux.go')
        cy.get('.file-actions')
    })

    it('should() - assert file tooltips are provided', () => {
        const elements = cy
            .contains('NewRouter')
            .get('span')
            .filter('.pl-en')
        const first = elements.first()
        first.scrollIntoView()
        first.trigger('mouseover')
        elements.first().click({ force: true })
        cy.contains('Go to definition')
    })

    it('should() - (Unified PR) assert base tooltips are provided', () => {
        cy.visit('https://github.com/gorilla/mux/pull/328/files?diff=unified')
        cy.contains(VIEW_FILE_BASE_BUTTON)
        cy.contains(VIEW_FILE_HEAD_BUTTON)
        const elements = cy
            .get('span')
            .filter('.pl-v')
            .contains('matchHost')
        const first = elements.first()
        first.scrollIntoView()
        first.trigger('mouseover')
        elements.first().click({ force: true })
        cy.contains('Go to definition')
    })

    it('should() - (Unified PR) assert head tooltips are provided', () => {
        cy.visit('https://github.com/gorilla/mux/pull/328/files?diff=unified')
        cy.contains(VIEW_FILE_BASE_BUTTON)
        cy.contains(VIEW_FILE_HEAD_BUTTON)
        const elements = cy
            .get('span')
            .filter('.pl-v')
            .contains('typ')
        const first = elements.first()
        first.scrollIntoView()
        first.trigger('mouseover')
        elements.first().click({ force: true })
        cy.contains('Go to definition')
    })

    it('should() - (Split PR) assert base tooltips are provided', () => {
        cy.visit('https://github.com/gorilla/mux/pull/328/files?diff=split')
        cy.contains(VIEW_FILE_BASE_BUTTON)
        cy.contains(VIEW_FILE_HEAD_BUTTON)
        const elements = cy
            .get('span')
            .filter('.pl-v')
            .contains('matchHost')
        const first = elements.first()
        first.scrollIntoView()
        first.trigger('mouseover')
        elements.first().click({ force: true })
        cy.contains('Go to definition')
    })

    it('should() - (Split PR) assert head tooltips are provided', () => {
        cy.visit('https://github.com/gorilla/mux/pull/328/files?diff=split')
        cy.contains(VIEW_FILE_BASE_BUTTON)
        cy.contains(VIEW_FILE_HEAD_BUTTON)
        const elements = cy
            .get('span')
            .filter('.pl-v')
            .contains('typ')
        const first = elements.first()
        first.scrollIntoView()
        first.trigger('mouseover')
        elements.first().click({ force: true })
        cy.contains('Go to definition')
    })

    it('should() - Assert diff expander elements are present', () => {
        cy.visit('https://github.com/aws/aws-sdk-go-v2/pull/189/files')
        cy.contains(VIEW_FILE_BASE_BUTTON)
        cy.get('.js-diff-load-container')
        cy.get('.js-diff-load-container')
            .first()
            .scrollIntoView()
        cy.get('.diff-table')
    })
})
