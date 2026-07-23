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
export type FunctionTag = { id: string; name: string; code: string; description?: string; publishedSpeciesCount?: number };
export type ApplicationCase = { id:string;industry:string;scenario:string;problem:string;solution:string;resultSummary:string;maturityLevel:string;source:string };
export type User = {id:string;email:string;displayName:string;role:string};
export type SearchHistory = {id:string;query:string;filters:Partial<SearchFilters>;resultCount:number;createdAt:string};
export type RecommendationEvidence = { id:string;title:string;publicationYear?:number;doi?:string;pmid?:string;sourceUrl?:string;conclusion:string;evidenceLevel:string;evidenceScore:number };
export type RecommendationItem = { id: string; slug: string; latinName: string; chineseName: string; safetyLevel: string; summary: string; score: number; evidenceCount: number; evidenceReferences: RecommendationEvidence[]; reasons: string[]; riskWarning?: string };
export type RecommendationResponse = { recordId: string; parsedFunctionTag?: string; parsedIntent?: {functionTag?:string;temperature?:number;ph?:number;safetyLevel?:string;sourceEnvironment?:string}; items: RecommendationItem[]; relaxationSuggestions?:string[]; disclaimer: string };
export type SearchFilters = { functionTag: string; temperature: string; ph: string; safetyLevel: string; sourceEnvironment: string };
export type SpeciesComparison = Species & { functionTags:string[];temperatureMin:number|null;temperatureMax:number|null;phMin:number|null;phMax:number|null;evidenceCount:number };
export type CombinationItem = {members:{id:string;slug:string;latinName:string;chineseName:string;safetyLevel:string;functionTags:string[];evidenceCount:number}[];score:number;temperatureMin?:number;temperatureMax?:number;phMin?:number;phMax?:number;compatible:boolean;validationStatus:'confirmed'|'contradicted'|'inconclusive'|'unverified';compatibleExperiments:number;incompatibleExperiments:number;reasons:string[];warning?:string};
export type CombinationResponse = { recordId:string;items:CombinationItem[];disclaimer:string };
