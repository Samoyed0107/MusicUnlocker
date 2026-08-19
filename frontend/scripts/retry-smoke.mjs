import assert from 'node:assert/strict'
import fs from 'node:fs'

const table = fs.readFileSync(new URL('../src/components/ProcessingTable.tsx', import.meta.url), 'utf8')
const app = fs.readFileSync(new URL('../src/App.tsx', import.meta.url), 'utf8')
const en = fs.readFileSync(new URL('../src/i18n/en-US.ts', import.meta.url), 'utf8')
const zh = fs.readFileSync(new URL('../src/i18n/zh-CN.ts', import.meta.url), 'utf8')
assert.match(table, /QMC_EXTERNAL_KEY_REQUIRED/)
assert.match(table, /KGG_DATABASE_REQUIRED/)
assert.match(table, /retryableExternalDatabase/)
assert.match(table, /retryableExternalKey/)
assert.match(table, /retryingIndex!==undefined/)
assert.match(table, /item\.status==='SUCCESS'/)
assert.match(app, /RetryQMCWithExternalKey/)
assert.match(app, /RetryKGGWithDatabase/)
for (const key of ['retry.provide', 'retry.another', 'retry.database', 'retry.anotherDatabase', 'retry.retrying', 'errors.qmcKeyFileInvalid']) {
  assert.match(en, new RegExp(`'${key}'`))
  assert.match(zh, new RegExp(`'${key}'`))
}
console.log('retry presentation smoke: PASS')
