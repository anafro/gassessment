import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';
import { URL } from 'node:url';
import ts from 'typescript';

// Execute the same TypeScript module used by Vue without a second test framework.
const source = readFileSync(new URL('./recovery.ts', import.meta.url), 'utf8');
const { outputText } = ts.transpileModule(source, {
    compilerOptions: {
        module: ts.ModuleKind.ESNext,
        target: ts.ScriptTarget.ES2020,
    },
});
const { recoverPending } = await import(
    `data:text/javascript,${encodeURIComponent(outputText)}`
);
const request = {
    key: 'original-request-key',
    match: 1,
    outcome: 0,
    amount: '10000',
    odds: 250,
};

test('restores original payload and key exactly', () => {
    assert.deepEqual(
        recoverPending(
            () => JSON.stringify(request),
            () => assert.fail('valid request discarded'),
        ),
        { pending: request, blocked: false },
    );
});
test('empty storage allows a new bet', () => {
    assert.deepEqual(
        recoverPending(
            () => null,
            () => assert.fail('empty storage discarded'),
        ),
        { blocked: false },
    );
});
test('storage access failure blocks new submissions', () => {
    assert.deepEqual(
        recoverPending(
            () => {
                throw new Error('denied');
            },
            () => assert.fail('unread storage discarded'),
        ),
        { blocked: true },
    );
});
test('corrupt or invalid payloads are discarded and allow a new bet', () => {
    for (const saved of [
        '',
        '{',
        'null',
        '[]',
        '{}',
        JSON.stringify({ ...request, amount: '01' }),
        JSON.stringify({ ...request, match: 7 }),
        JSON.stringify({ ...request, outcome: -1 }),
        JSON.stringify({ ...request, odds: 1.5 }),
        JSON.stringify({ ...request, key: 'short' }),
        JSON.stringify({ ...request, extra: true }),
    ]) {
        let discarded = false;
        assert.deepEqual(
            recoverPending(
                () => saved,
                () => {
                    discarded = true;
                },
            ),
            { blocked: false },
            saved,
        );
        assert.equal(discarded, true);
    }
});

test('failed removal blocks new submissions', () => {
    assert.deepEqual(
        recoverPending(
            () => '{',
            () => {
                throw new Error('denied');
            },
        ),
        { blocked: true },
    );
});
