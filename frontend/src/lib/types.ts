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

export interface AnalyticsSnapshot {
  formId: string;
  count: number;
  fields: Record<string, any>;
}

export interface SubmitResponseBody {
  answers: Record<string, unknown>;
}
