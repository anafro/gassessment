/** Server line snapshot; odds are integer hundredths in P1/draw/P2 order. */
export type Match = {
    id: number;
    name: string;
    odds: number[];
};
/** Frozen price chosen by the player, independent of subsequent live updates. */
export type Selection = {
    match: number;
    outcome: number;
    odds: number;
};
