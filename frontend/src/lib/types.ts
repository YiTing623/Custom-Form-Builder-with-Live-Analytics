export type FieldType = "text" | "multiple" | "checkbox" | "rating";

export interface FormField {
  id: string;
  type: FieldType;
  label: string;
  required?: boolean;
  options?: string[];
  max?: number;
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
