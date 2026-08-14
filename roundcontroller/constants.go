package roundcontroller

// Round rewards. What steers behaviour is not the absolute values but the
// two gaps between them: WIN-DRAW (+2 here) is what a win is worth over
// settling for a draw, and DRAW-LOOSE (+6) is what not losing is worth.
//
// The heavy loss penalty is deliberate and was checked by experiment.
// Raising WIN to 6 to close that gap makes the agent worse at everything,
// including at converting wins: three runs at 100 generations went from
// 15% to 29% losses and from 59% to 49% of immediate wins taken. Both
// sides share these rewards, so a bigger win bonus mostly makes the
// opposing X population gamble harder, and neither side ever consolidates
// blocking - which fell from 80% to 59%. Blocking is the foundation here:
// an agent that cannot defend never reaches a position worth winning.
const WIN_SCORE = 3
const DRAW_SCORE = 1
const LOOSE_SCORE = -5
