/** Display minor units or odds hundredths; never use this for balance math. */
export const money = (n: number) => (n / 100).toFixed(2);
