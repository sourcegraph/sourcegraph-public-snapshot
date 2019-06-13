import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { ThreadSettings } from '../../../../../web/src/enterprise/threads/settings'
import { FeatureProviderRegistry } from './registry'

export interface CheckTemplate {
    id: string
    title: string
    description?: string
    icon?: any
    iconColor?: string
    settings?: ThreadSettings
}

export interface CheckTemplateRegistrationOptions {}

export type ProvideCheckTemplateSignature = CheckTemplate

/** Provides check templates from all extensions. */
export class CheckTemplateRegistry extends FeatureProviderRegistry<
    CheckTemplateRegistrationOptions,
    ProvideCheckTemplateSignature
> {
    /**
     * Returns an observable that emits the specified check template whenever it or the set of
     * registered check templates changes.
     */
    public getCheckTemplate(id: string): Observable<CheckTemplate | null> {
        return this.getCheckTemplates().pipe(map(checkTemplates => checkTemplates.find(t => t.id === id) || null))
    }

    /**
     * Returns an observable that emits all check templates whenever the registered set or any item's properties
     * change.
     */
    public getCheckTemplates(): Observable<CheckTemplate[]> {
        return this.entries.pipe(map(e => e.map(e => e.provider)))
    }
}
