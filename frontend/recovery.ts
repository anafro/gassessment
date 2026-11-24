/** A frozen bet payload. Recovery must never replace its idempotency key. */
export type BetRequest = {
    key: string;
    match: number;
    outcome: number;
    amount: string;
    odds: number;
};

export const pendingStorageKey = 'sportsbook.pending.v1';

/** Validate persisted data before it reaches the template or a retry request. */
function isBetRequest(value: unknown): value is BetRequest {
    if (typeof value !== 'object' || value === null || Array.isArray(value))
        return false;
    const fields = value as Record<string, unknown>;
    return (
        Object.keys(fields).length === 5 &&
        typeof fields.key === 'string' &&
        new TextEncoder().encode(fields.key).length >= 16 &&
        new TextEncoder().encode(fields.key).length <= 100 &&
        typeof fields.amount === 'string' &&
        /^[1-9]\d{0,8}$/.test(fields.amount) &&
        Number(fields.amount) <= 100000000 &&
        typeof fields.match === 'number' &&
        Number.isInteger(fields.match) &&
        fields.match >= 1 &&
        fields.match <= 6 &&
        typeof fields.outcome === 'number' &&
        Number.isInteger(fields.outcome) &&
        fields.outcome >= 0 &&
        fields.outcome <= 2 &&
        typeof fields.odds === 'number' &&
        Number.isInteger(fields.odds) &&
        fields.odds >= 101 &&
        fields.odds <= 10000
    );
}

/** Discard corrupt payloads; inaccessible storage still blocks new submissions. */
export function recoverPending(
    read: () => string | null,
    discard: () => void,
): {
    pending?: BetRequest;
    blocked: boolean;
} {
    let saved: string | null;
    try {
        saved = read();
    } catch {
        return { blocked: true };
    }
    if (saved === null) return { blocked: false };
    try {
        const parsed: unknown = JSON.parse(saved);
        if (isBetRequest(parsed)) return { pending: parsed, blocked: false };
    } catch {
        // Invalid JSON is discarded just like a structurally invalid request.
    }
    try {
        discard();
        return { blocked: false };
    } catch {
        return { blocked: true };
    }
}
