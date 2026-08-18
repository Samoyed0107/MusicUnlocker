import { createContext, useContext, useMemo, useState, type ReactNode } from 'react'
import { enUS } from './en-US'
import { zhCN } from './zh-CN'
import { interpolationParity, interpolationVariables } from './validation'

export type TranslationResource = { [K in keyof typeof enUS]: string }
export type LanguagePreference = 'SYSTEM' | 'zh-CN' | 'en-US'
export type TranslationKey = keyof typeof enUS
type Locale = 'en-US' | 'zh-CN'
export const resources: Record<Locale, TranslationResource> = { 'en-US': enUS, 'zh-CN': zhCN }
export { interpolationParity, interpolationVariables }
export function resolveLocale(preference: LanguagePreference, languages?: readonly string[]): Locale { if (preference === 'zh-CN' || preference === 'en-US') return preference; const availableLanguages = languages ?? (typeof navigator === 'undefined' ? [] : navigator.languages ?? [navigator.language]); const value = (availableLanguages[0] ?? '').toLowerCase().replace('_', '-'); if (value === 'zh-hans' || value === 'zh-cn' || value === 'zh-sg' || value.startsWith('zh-hans-')) return 'zh-CN'; return 'en-US' }
export function normalizeLanguage(value: unknown): LanguagePreference { return value === 'zh-CN' || value === 'en-US' || value === 'SYSTEM' ? value : 'SYSTEM' }
export function validateResources() { const keys = Object.keys(enUS).sort(); const other = Object.keys(zhCN).sort(); if (keys.length !== other.length || keys.some((key, i) => key !== other[i])) throw new Error('Localization resource key mismatch'); if (Object.values(enUS).some(value => !value)) throw new Error('English localization contains an empty value'); const mismatches = interpolationParity(zhCN, enUS); if (mismatches.length) throw new Error(`Localization interpolation mismatch: ${mismatches[0]}`); if (resolveLocale('SYSTEM', ['zh-SG']) !== 'zh-CN' || resolveLocale('SYSTEM', ['zh-Hans']) !== 'zh-CN' || resolveLocale('SYSTEM', ['zh-TW']) !== 'en-US' || resolveLocale('SYSTEM', ['fr-FR']) !== 'en-US' || normalizeLanguage('invalid') !== 'SYSTEM') throw new Error('Localization locale policy validation failed') }
function reportFallback(locale: Locale, key: TranslationKey) { if (import.meta.env?.DEV) console.error(`Localization fallback: ${locale}.${key}`) }
export function translate(locale: Locale, key: TranslationKey, vars: Record<string, string | number> = {}) { const selected = resources[locale][key]; const fallback = resources['en-US'][key]; const template = locale !== 'en-US' && interpolationVariables(selected ?? '').join(',') !== interpolationVariables(fallback).join(',') ? (reportFallback(locale, key), fallback) : selected ?? fallback; return Object.entries(vars).reduce((text, [name, value]) => text.replaceAll(`{{${name}}}`, String(value)), template) }
export const LocalizationContext = createContext<{ locale: Locale; preference: LanguagePreference; setPreference: (value: LanguagePreference) => void; t: (key: TranslationKey, vars?: Record<string, string | number>) => string }>({ locale: 'en-US', preference: 'SYSTEM', setPreference: () => {}, t: (key, vars) => translate('en-US', key, vars) })
export const useTranslation = () => useContext(LocalizationContext)
export function LocalizationProvider({ children }: { children: ReactNode }) { const [preference, setPreference] = useState<LanguagePreference>('SYSTEM'); const locale = resolveLocale(preference); const value = useMemo(() => ({ locale, preference, setPreference, t: (key: TranslationKey, vars?: Record<string, string | number>) => translate(locale, key, vars) }), [locale, preference]); return <LocalizationContext.Provider value={value}>{children}</LocalizationContext.Provider> }
