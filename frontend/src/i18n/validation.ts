export function interpolationVariables(value: string): string[] {
  return [...value.matchAll(/\{\{(\w+)\}\}/g)].map(match => match[1]).sort()
}

export function interpolationParity(resource: Record<string, string>, fallback: Record<string, string>): string[] {
  return Object.keys(fallback).filter(key => interpolationVariables(fallback[key]).join(',') !== interpolationVariables(resource[key] ?? '').join(','))
}
