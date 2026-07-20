export type SpeciesStatus = 'draft' | 'pending_review' | 'published' | 'archived';

export type Species = {
  id: string;
  slug: string;
  latinName: string;
  chineseName: string;
  strainNumber: string;
  sourceEnvironment: string;
  safetyLevel: string;
  isModelOrganism: boolean;
  summary: string;
  status: SpeciesStatus;
  dataQualityScore: number;
  createdAt: string;
  updatedAt: string;
  publishedAt?: string;
};

export type ListResponse = { items: Species[]; total: number; limit: number; offset: number };
export type SpeciesFunction = { functionTagId: string; functionTagName: string };
export type CultureCondition = { id: string; mediumName: string; temperatureMin: number | null; temperatureMax: number | null; phMin: number | null; phMax: number | null; oxygenRequirement: string; cultureTime: string; notes: string };
export type Evidence = { id: string; title: string; authors: string; journal: string; publicationYear: number | null; doi: string; pmid: string; sourceUrl: string; conclusion: string; evidenceLevel: string; evidenceScore: number };
export type SpeciesAlias = { id: string; name: string; type: string; source: string };
export type FunctionTag = { id: string; name: string; code: string };
export type RecommendationItem = { id: string; slug: string; latinName: string; chineseName: string; safetyLevel: string; summary: string; score: number; evidenceCount: number; reasons: string[]; riskWarning?: string };
export type RecommendationResponse = { recordId: string; parsedFunctionTag?: string; items: RecommendationItem[]; disclaimer: string };
export type SearchFilters = { functionTag: string; temperature: string; ph: string; safetyLevel: string; sourceEnvironment: string };
