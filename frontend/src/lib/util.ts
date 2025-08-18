export function uid(prefix = "f"): string {
  return `${prefix}-${Math.random().toString(36).slice(2, 10)}`;
}