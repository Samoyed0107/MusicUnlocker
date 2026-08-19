import assert from 'node:assert/strict'
import fs from 'node:fs'

const source = fs.readFileSync(new URL('../src/components/AudioPreview.tsx', import.meta.url), 'utf8')
assert.equal((source.match(/GetPreviewURL/g) ?? []).length, 1)
assert.match(source, /onErrorRef\.current\(String\(error\)\)/)
assert.match(source, /\}, \[item\?\.index, batchId, api\]\)/)
assert.doesNotMatch(source, /\}, \[item\?\.index, batchId, api, onError\]\)/)
console.log('preview continuity smoke: PASS')
