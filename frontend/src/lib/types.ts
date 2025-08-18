export type FieldType = "text" | "multiple" | "checkbox" | "rating";

export type ConditionOperator = "eq" | "ne" | "includes" | "gt" | "gte" | "lt" | "lte";

export interface ShowIf {
  fieldId: string;
  op: ConditionOperator;
  value: any;
}

export interface FormField {
  id: string;
  type: FieldType;
  label: string;
  required?: boolean;
  options?: string[];
  max?: number;
  showIf?: ShowIf;
}

export interface FormDoc {
  id: string;
  title: string;
  fields: FormField[];
  status: "draft" | "published";
}

export interface Trends {
  avgRating?: number;
  mostCommon?: Record<string, string | string[]>;
  skipped?: Record<string, number>;
  mostSkipped?: { id: string; label: string; skipped: number; total: number }[];
}

export interface AnalyticsSnapshot {
  formId: string;
  count: number;
  fields: Record<
    string,
    | { type: "text"; nonEmptyCount: number }
    | { type: "multiple"; distribution: Record<string, number> }
    | { type: "checkbox"; distribution: Record<string, number> }
    | { type: "rating"; distribution: Record<number, number>; average: number }
  >;
  trends?: Trends;
}

export interface SubmitResponseBody {
  answers: Record<string, unknown>;
}
