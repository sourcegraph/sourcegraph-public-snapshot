import { getCombinedViews, getInsightCombinedViews } from './api/get-combined-views';
import { getSubjectSettings, updateSubjectSettings } from './api/subject-settings';
import { ApiService } from './types'

/**
 * Main API service to get data for code insights
 * */
export const createInsightAPI = (): ApiService => ({
    getCombinedViews,
    getInsightCombinedViews,
    getSubjectSettings,
    updateSubjectSettings,
})
