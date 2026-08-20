import assert from 'node:assert/strict'
import fs from 'node:fs'

const theme = fs.readFileSync(new URL('../src/theme.ts', import.meta.url), 'utf8')
const app = fs.readFileSync(new URL('../src/App.tsx', import.meta.url), 'utf8')
const about = fs.readFileSync(new URL('../src/components/AboutDialog.tsx', import.meta.url), 'utf8')
const badge = fs.readFileSync(new URL('../src/components/StatusBadge.tsx', import.meta.url), 'utf8')

for (const value of ['SYSTEM', 'LIGHT', 'DARK']) assert.match(theme, new RegExp(`'${value}'`))
assert.match(app, /GetPreferences/)
assert.match(app, /SavePreferences/)
assert.match(app, /SetWindowTheme/)
assert.match(about, /appInfo\.version/)
assert.doesNotMatch(about, /0\.\d+\.\d+/)
assert.match(about, /about\.supportedFormats/)
assert.match(about, /about\.offline/)
assert.match(badge, /<Icon name=/)
assert.doesNotMatch(badge, /symbol: ['"`][✓!×–?•]/)
console.log('theme/about smoke: PASS')
