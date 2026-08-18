import assert from 'node:assert/strict'
import fs from 'node:fs'
import { interpolationParity, interpolationVariables } from '../src/i18n/validation.ts'

const readResource = file => {
  const source = fs.readFileSync(new URL(file, import.meta.url), 'utf8')
  return Object.fromEntries([...source.matchAll(/'([^']+)':\s*'([^']*)'/g)].map(match => [match[1], match[2]]))
}

const enUS = readResource('../src/i18n/en-US.ts')
const zhCN = readResource('../src/i18n/zh-CN.ts')
assert.equal(enUS['batch.itemsAdded'], '{{count}} items added from Windows.')
assert.deepEqual(interpolationVariables(enUS['batch.itemsAdded']), ['count'])
assert.deepEqual(interpolationVariables(zhCN['batch.itemsAdded']), ['count'])
assert.deepEqual(interpolationParity({ ...zhCN, 'batch.itemsAdded': '{{count}} items added from Windows.' }, enUS), [])
assert.deepEqual(interpolationParity({ ...zhCN, 'batch.itemsAdded': '已添加 {{count}} {{plural}} 个项目。' }, enUS), ['batch.itemsAdded'])
assert.deepEqual(Object.keys(enUS).sort(), Object.keys(zhCN).sort())
console.log('i18n smoke: PASS')
